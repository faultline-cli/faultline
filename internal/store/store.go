package store

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"faultline/internal/model"
)

type Mode string

const (
	ModeAuto Mode = "auto"
	ModeOff  Mode = "off"
)

const (
	defaultStoreSubdir = ".faultline"
	defaultStoreFile   = "store.db"
)

type Config struct {
	Mode   Mode
	Path   string
	Strict bool
}

type Info struct {
	Mode     Mode
	Backend  string
	Path     string
	Degraded bool
	Warning  string
}

type Store interface {
	BeginRun(context.Context, BeginRunParams) (RunHandle, error)
	CompleteRun(context.Context, RunHandle, CompleteRunParams) error
	RecordWorkflowExecution(context.Context, *model.WorkflowExecutionRecord) (*model.WorkflowExecutionRecord, error)
	GetWorkflowExecution(context.Context, string) (*model.WorkflowExecutionRecord, error)
	ListWorkflowExecutions(context.Context, int) ([]model.WorkflowExecutionSummary, error)
	LookupSignatureHistory(context.Context, string) (SignatureHistory, error)
	CountSeenFailure(context.Context, string) (int, error)
	RecentTopFailures(context.Context, int) ([]string, error)
	ListSignatures(context.Context, int) ([]SignatureSummary, error)
	GetRecentFindingsBySignature(context.Context, string, int) ([]FindingSummary, error)
	ListPlaybookStats(context.Context, int) ([]PlaybookStats, error)
	LookupHookHistory(context.Context, string, string) (*HookHistorySummary, error)
	ListHookStats(context.Context, int) ([]HookStats, error)
	VerifyDeterminismForInputHash(context.Context, string) (DeterminismSummary, error)
	Close() error
}

type BeginRunParams struct {
	Surface    string
	SourceKind string
	Source     string
	InputHash  string
	StartedAt  time.Time
}

type RunHandle struct {
	ID int64
}

type CompleteRunParams struct {
	CompletedAt time.Time
	Analysis    *model.Analysis
}

type SignatureHistory struct {
	SignatureHash   string
	SeenBefore      bool
	OccurrenceCount int
	FirstSeenAt     string
	LastSeenAt      string
}

type FindingSummary struct {
	RunID         int64
	FailureID     string
	Title         string
	Category      string
	SignatureHash string
	SeenAt        string
}

type SignatureSummary struct {
	SignatureHash   string `json:"signature_hash,omitempty"`
	FailureID       string `json:"failure_id,omitempty"`
	Title           string `json:"title,omitempty"`
	Category        string `json:"category,omitempty"`
	OccurrenceCount int    `json:"occurrence_count,omitempty"`
	FirstSeenAt     string `json:"first_seen_at,omitempty"`
	LastSeenAt      string `json:"last_seen_at,omitempty"`
}

type PlaybookStats struct {
	FailureID           string  `json:"failure_id,omitempty"`
	Title               string  `json:"title,omitempty"`
	Category            string  `json:"category,omitempty"`
	SelectedCount       int     `json:"selected_count,omitempty"`
	MatchedCount        int     `json:"matched_count,omitempty"`
	NonSelectedCount    int     `json:"non_selected_count,omitempty"`
	AvgRank             float64 `json:"avg_rank,omitempty"`
	RecurringRunCount   int     `json:"recurring_run_count,omitempty"`
	RecurringSignatures int     `json:"recurring_signatures,omitempty"`
	AvgConfidence       float64 `json:"avg_confidence,omitempty"`
	LastSeenAt          string  `json:"last_seen_at,omitempty"`
}

type HookHistorySummary struct {
	TotalCount    int    `json:"total_count,omitempty"`
	ExecutedCount int    `json:"executed_count,omitempty"`
	PassedCount   int    `json:"passed_count,omitempty"`
	FailedCount   int    `json:"failed_count,omitempty"`
	BlockedCount  int    `json:"blocked_count,omitempty"`
	SkippedCount  int    `json:"skipped_count,omitempty"`
	LastSeenAt    string `json:"last_seen_at,omitempty"`
}

type HookStats struct {
	PlaybookID         string  `json:"playbook_id,omitempty"`
	HookID             string  `json:"hook_id,omitempty"`
	Category           string  `json:"category,omitempty"`
	TotalCount         int     `json:"total_count,omitempty"`
	ExecutedCount      int     `json:"executed_count,omitempty"`
	PassedCount        int     `json:"passed_count,omitempty"`
	FailedCount        int     `json:"failed_count,omitempty"`
	BlockedCount       int     `json:"blocked_count,omitempty"`
	SkippedCount       int     `json:"skipped_count,omitempty"`
	AvgConfidenceDelta float64 `json:"avg_confidence_delta,omitempty"`
	LastSeenAt         string  `json:"last_seen_at,omitempty"`
}

type DeterminismSummary struct {
	RunCount             int    `json:"run_count,omitempty"`
	DistinctOutputHashes int    `json:"distinct_output_hashes,omitempty"`
	FirstSeenAt          string `json:"first_seen_at,omitempty"`
	LastSeenAt           string `json:"last_seen_at,omitempty"`
	Stable               bool   `json:"stable,omitempty"`
}

func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(home, defaultStoreSubdir, defaultStoreFile), nil
}

func ResolveConfig(raw string, disable bool) (Config, error) {
	if disable {
		return Config{Mode: ModeOff}, nil
	}
	value := strings.TrimSpace(raw)
	switch {
	case value == "", strings.EqualFold(value, string(ModeAuto)):
		return Config{Mode: ModeAuto}, nil
	case strings.EqualFold(value, string(ModeOff)):
		return Config{Mode: ModeOff}, nil
	default:
		return Config{Mode: ModeAuto, Path: value}, nil
	}
}

func OpenBestEffort(cfg Config) (Store, Info, error) {
	if cfg.Mode == "" {
		cfg.Mode = ModeAuto
	}
	if cfg.Mode == ModeOff {
		return Noop(), Info{Mode: ModeOff}, nil
	}
	path := strings.TrimSpace(cfg.Path)
	if path == "" {
		var err error
		path, err = DefaultPath()
		if err != nil {
			if cfg.Strict {
				return nil, Info{}, err
			}
			return Noop(), Info{
				Mode:     ModeOff,
				Degraded: true,
				Warning:  err.Error(),
			}, nil
		}
	}
	sqliteStore, err := openSQLite(path)
	if err != nil {
		if cfg.Strict {
			return nil, Info{}, err
		}
		return Noop(), Info{
			Mode:     ModeOff,
			Path:     path,
			Backend:  "sqlite",
			Degraded: true,
			Warning:  err.Error(),
		}, nil
	}
	return sqliteStore, Info{
		Mode:    ModeAuto,
		Path:    path,
		Backend: "sqlite",
	}, nil
}
