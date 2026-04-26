package model

type HookCategory string

const (
	HookCategoryVerify    HookCategory = "verify"
	HookCategoryCollect   HookCategory = "collect"
	HookCategoryRemediate HookCategory = "remediate"
)

type HookMode string

const (
	HookModeOff         HookMode = "off"
	HookModeVerifyOnly  HookMode = "verify-only"
	HookModeCollectOnly HookMode = "collect-only"
	HookModeSafe        HookMode = "safe"
	HookModeFull        HookMode = "full"
)

type HookKind string

const (
	HookKindFileExists           HookKind = "file_exists"
	HookKindDirExists            HookKind = "dir_exists"
	HookKindEnvVarPresent        HookKind = "env_var_present"
	HookKindCommandExitZero      HookKind = "command_exit_zero"
	HookKindCommandOutputMatches HookKind = "command_output_matches"
	HookKindCommandOutputCapture HookKind = "command_output_capture"
	HookKindReadFileExcerpt      HookKind = "read_file_excerpt"
)

type PlaybookHooks struct {
	Verify    []HookDefinition `yaml:"verify,omitempty" json:"verify,omitempty"`
	Collect   []HookDefinition `yaml:"collect,omitempty" json:"collect,omitempty"`
	Remediate []HookDefinition `yaml:"remediate,omitempty" json:"remediate,omitempty"`
	Disable   []string         `yaml:"disable,omitempty" json:"disable,omitempty"`
}

type HookDefinition struct {
	ID              string   `yaml:"id,omitempty" json:"id,omitempty"`
	Use             string   `yaml:"use,omitempty" json:"use,omitempty"`
	Extends         string   `yaml:"extends,omitempty" json:"extends,omitempty"`
	Kind            HookKind `yaml:"kind,omitempty" json:"kind,omitempty"`
	Path            string   `yaml:"path,omitempty" json:"path,omitempty"`
	EnvVar          string   `yaml:"env_var,omitempty" json:"env_var,omitempty"`
	Command         []string `yaml:"command,omitempty" json:"command,omitempty"`
	Pattern         string   `yaml:"pattern,omitempty" json:"pattern,omitempty"`
	Lines           int      `yaml:"lines,omitempty" json:"lines,omitempty"`
	MaxBytes        int      `yaml:"max_bytes,omitempty" json:"max_bytes,omitempty"`
	ConfidenceDelta float64  `yaml:"confidence_delta,omitempty" json:"confidence_delta,omitempty"`
	Metadata        HookMeta `yaml:"-" json:"metadata,omitempty"`
}

type HookMeta struct {
	SourcePack string `json:"source_pack,omitempty"`
	SourceFile string `json:"source_file,omitempty"`
}

type HookStatus string

const (
	HookStatusExecuted HookStatus = "executed"
	HookStatusSkipped  HookStatus = "skipped"
	HookStatusBlocked  HookStatus = "blocked"
	HookStatusFailed   HookStatus = "failed"
)

type HookFact struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type HookResult struct {
	ID              string       `json:"id"`
	Category        HookCategory `json:"category"`
	Kind            HookKind     `json:"kind,omitempty"`
	Status          HookStatus   `json:"status"`
	Passed          *bool        `json:"passed,omitempty"`
	ConfidenceDelta float64      `json:"confidence_delta,omitempty"`
	Reason          string       `json:"reason,omitempty"`
	Facts           []HookFact   `json:"facts,omitempty"`
	Evidence        []string     `json:"evidence,omitempty"`
	SourcePack      string       `json:"source_pack,omitempty"`
	SourceFile      string       `json:"source_file,omitempty"`
}

type HookReport struct {
	Mode            HookMode     `json:"mode,omitempty"`
	BaseConfidence  float64      `json:"base_confidence,omitempty"`
	ConfidenceDelta float64      `json:"confidence_delta,omitempty"`
	FinalConfidence float64      `json:"final_confidence,omitempty"`
	Results         []HookResult `json:"results,omitempty"`
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
