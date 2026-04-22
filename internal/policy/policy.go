// Package policy computes deterministic advisory quarantine policy from
// pipeline reliability metrics. All recommendations are advisory only.
// Faultline never triggers retries, test routing, or CI orchestration.
//
// Recommendation levels (in increasing urgency):
//   - "ok":         metrics look healthy or there is insufficient history.
//   - "observe":    a recurring pattern is emerging (TSS ≥ 0.3).
//   - "quarantine": persistent recurrence (TSS ≥ 0.5) or low pipeline
//                   health (FPC < 0.5 or PHI < 0.4).
//   - "blocking":   high-confidence persistent critical/high-severity failure
//                   (TSS ≥ 0.8 and severity is "critical" or "high").
//
// Absent metrics yield an absent (nil) policy — Faultline does not invent
// recommendations when the required data is not available.
package policy

import (
	"fmt"

	"faultline/internal/model"
)

const (
	// Recommendation strings. These are the stable values exposed in the
	// public JSON artifact.
	RecommendOK         = "ok"
	RecommendObserve    = "observe"
	RecommendQuarantine = "quarantine"
	RecommendBlocking   = "blocking"

	// Thresholds are documented so external automation can gate on them.

	// tssBlockingThreshold: TSS at or above this level combined with
	// critical/high severity → blocking recommendation.
	tssBlockingThreshold = 0.8

	// tssQuarantineThreshold: TSS at or above this level → quarantine.
	tssQuarantineThreshold = 0.5

	// tssObserveThreshold: TSS at or above this level → observe.
	tssObserveThreshold = 0.3

	// fpcQuarantineThreshold: FPC below this level → quarantine.
	fpcQuarantineThreshold = 0.5

	// phiQuarantineThreshold: PHI below this level → quarantine.
	phiQuarantineThreshold = 0.4
)

// Compute derives an advisory policy recommendation from m and the top result
// severity. Returns nil when m is nil so the policy field stays absent from
// output when no metrics are available.
func Compute(m *model.Metrics, topSeverity string) *model.Policy {
	if m == nil {
		return nil
	}

	// If none of the metric pointers are set, there is nothing to act on.
	if m.TSS == nil && m.FPC == nil && m.PHI == nil {
		return nil
	}

	recommendation, reason, basis := evaluate(m, topSeverity)
	return &model.Policy{
		Recommendation: recommendation,
		Reason:         reason,
		Basis:          basis,
	}
}

// evaluate applies the threshold rules in priority order and returns the
// recommendation, reason, and metric basis.
func evaluate(m *model.Metrics, severity string) (string, string, []string) {
	var basis []string

	isCritical := severity == "critical" || severity == "high"

	// --- blocking ---
	if m.TSS != nil && *m.TSS >= tssBlockingThreshold && isCritical {
		basis = append(basis, "tss")
		return RecommendBlocking,
			fmt.Sprintf("persistent critical failure: TSS %.2f with %s severity — recommend halting until root cause is resolved", *m.TSS, severity),
			basis
	}

	// --- quarantine ---
	quarantineReasons := []string{}
	if m.TSS != nil && *m.TSS >= tssQuarantineThreshold {
		basis = append(basis, "tss")
		quarantineReasons = append(quarantineReasons, fmt.Sprintf("TSS %.2f (recurring pattern)", *m.TSS))
	}
	if m.FPC != nil && *m.FPC < fpcQuarantineThreshold {
		basis = append(basis, "fpc")
		quarantineReasons = append(quarantineReasons, fmt.Sprintf("FPC %.2f (low pattern coverage)", *m.FPC))
	}
	if m.PHI != nil && *m.PHI < phiQuarantineThreshold {
		basis = append(basis, "phi")
		quarantineReasons = append(quarantineReasons, fmt.Sprintf("PHI %.2f (degraded pipeline health)", *m.PHI))
	}
	if len(quarantineReasons) > 0 {
		return RecommendQuarantine,
			"isolate and review: " + join(quarantineReasons),
			dedupe(basis)
	}

	// --- observe ---
	if m.TSS != nil && *m.TSS >= tssObserveThreshold {
		return RecommendObserve,
			fmt.Sprintf("pattern emerging: TSS %.2f — monitor for continued recurrence", *m.TSS),
			[]string{"tss"}
	}
	if m.PHI != nil && *m.PHI < 0.6 {
		return RecommendObserve,
			fmt.Sprintf("pipeline health degrading: PHI %.2f — review failure mix", *m.PHI),
			[]string{"phi"}
	}

	// --- ok ---
	return RecommendOK, "metrics within acceptable range", nil
}

func join(parts []string) string {
	out := ""
	for i, p := range parts {
		if i > 0 {
			out += "; "
		}
		out += p
	}
	return out
}

func dedupe(s []string) []string {
	seen := map[string]struct{}{}
	out := s[:0:len(s)]
	for _, v := range s {
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			out = append(out, v)
		}
	}
	return out
}
