package model

type HypothesisSpec struct {
	Supports       []HypothesisSignal        `yaml:"supports,omitempty" json:"supports,omitempty"`
	Contradicts    []HypothesisSignal        `yaml:"contradicts,omitempty" json:"contradicts,omitempty"`
	Discriminators []HypothesisDiscriminator `yaml:"discriminators,omitempty" json:"discriminators,omitempty"`
	Excludes       []HypothesisSignal        `yaml:"excludes,omitempty" json:"excludes,omitempty"`
}

type HypothesisSignal struct {
	Signal string  `yaml:"signal,omitempty" json:"signal,omitempty"`
	Weight float64 `yaml:"weight,omitempty" json:"weight,omitempty"`
}

type HypothesisDiscriminator struct {
	Description string  `yaml:"description,omitempty" json:"description,omitempty"`
	Signal      string  `yaml:"signal,omitempty" json:"signal,omitempty"`
	Weight      float64 `yaml:"weight,omitempty" json:"weight,omitempty"`
}

type HypothesisMatch struct {
	Signal      string   `json:"signal,omitempty"`
	Description string   `json:"description,omitempty"`
	Weight      float64  `json:"weight,omitempty"`
	Evidence    []string `json:"evidence,omitempty"`
}

type HypothesisAssessment struct {
	BaseScore      float64           `json:"base_score,omitempty"`
	FinalScore     float64           `json:"final_score,omitempty"`
	Eliminated     bool              `json:"eliminated,omitempty"`
	Supports       []HypothesisMatch `json:"supports,omitempty"`
	Contradicts    []HypothesisMatch `json:"contradicts,omitempty"`
	Discriminators []HypothesisMatch `json:"discriminators,omitempty"`
	Excludes       []HypothesisMatch `json:"excludes,omitempty"`
	Why            []string          `json:"why,omitempty"`
	WhyLessLikely  []string          `json:"why_less_likely,omitempty"`
	RuledOutBy     []string          `json:"ruled_out_by,omitempty"`
	DisproofChecks []string          `json:"disproof_checks,omitempty"`
}
