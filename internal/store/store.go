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
	GetRecentFindingsBySignature(context.Context, string, int) ([]FindingSummary, error)
	LookupHookHistory(context.Context, string, string) (*HookHistorySummary, error)
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

type HookHistorySummary struct {
	TotalCount    int    `json:"total_count,omitempty"`
	ExecutedCount int    `json:"executed_count,omitempty"`
	PassedCount   int    `json:"passed_count,omitempty"`
	FailedCount   int    `json:"failed_count,omitempty"`
	BlockedCount  int    `json:"blocked_count,omitempty"`
	SkippedCount  int    `json:"skipped_count,omitempty"`
	LastSeenAt    string `json:"last_seen_at,omitempty"`
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
