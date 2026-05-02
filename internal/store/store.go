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
	LookupSignatureHistory(context.Context, string) (SignatureHistory, error)
	CountSeenFailure(context.Context, string) (int, error)
	RecentTopFailures(context.Context, int) ([]string, error)
	ListFailureReports(context.Context, int) ([]FailureReport, error)
	ListSignatures(context.Context, int) ([]SignatureSummary, error)
	GetRecentFindingsBySignature(context.Context, string, int) ([]FindingSummary, error)
	ListPlaybookStats(context.Context, int) ([]PlaybookStats, error)
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

type FailureReport struct {
	FailureID       string `json:"failure_id,omitempty"`
	Count           int    `json:"count"`
	LastSeenAt      string `json:"last_seen_at,omitempty"`
	ExampleEvidence string `json:"example_evidence,omitempty"`
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

func ResolveConfig(raw string) (Config, error) {
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
