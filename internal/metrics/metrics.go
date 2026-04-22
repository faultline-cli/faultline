// Package metrics computes pipeline reliability metrics from analysis history.
//
// Three metrics are defined:
//
//   - TSS (Trace Stability Score): fraction of locally-recorded analysis runs
//     where the same failure pattern appeared. Computed from the local history
//     store; absent when fewer than 2 matched entries exist.
//
//   - FPC (Failure Pattern Coverage): fraction of all runs in a supplied
//     history file that matched a known playbook. Absent when the file
//     contains fewer than 3 entries.
//
//   - PHI (Pipeline Health Index): composite score that weighs FPC against
//     the degree to which a single failure pattern dominates matched runs.
//     Absent when the history file contains fewer than 5 entries.
//
// All values are rounded to two decimal places. Absent data yields absent
// fields — Faultline never invents values.
package metrics

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"sort"

	"faultline/internal/model"
)

const (
	// minTSSEntries is the minimum number of local history entries required
	// to emit a TSS value.
	minTSSEntries = 2
	// minFPCEntries is the minimum number of explicit history entries
	// required to emit an FPC value.
	minFPCEntries = 3
	// minPHIEntries is the minimum number of explicit history entries
	// required to emit a PHI value.
	minPHIEntries = 5

	// tssWarnThreshold is the TSS level below which a drift component is emitted.
	tssWarnThreshold = 0.5
	// fpcWarnThreshold is the FPC level below which a drift component is emitted.
	fpcWarnThreshold = 0.7
	// phiWarnThreshold is the PHI level below which a drift component is emitted.
	phiWarnThreshold = 0.5
)

// LocalEntry is a single entry from the local history store consumed by TSS.
type LocalEntry struct {
	FailureID string
}

// FromLocalHistory computes TSS for currentFailureID using the supplied local
// history entries. Entries should be the most-recent matched runs from the
// local history store. Returns nil when fewer than minTSSEntries are supplied.
func FromLocalHistory(currentFailureID string, entries []LocalEntry) *model.Metrics {
	n := len(entries)
	if n < minTSSEntries {
		return nil
	}

	seen := 0
	for _, e := range entries {
		if e.FailureID == currentFailureID {
			seen++
		}
	}

	tss := round2(float64(seen) / float64(n))
	m := &model.Metrics{
		TSS:          &tss,
		HistoryCount: n,
	}

	if tss >= tssWarnThreshold && seen > 0 {
		m.DriftComponents = append(m.DriftComponents,
			fmt.Sprintf("persistent failure: %s is %.0f%% of recent analyzed runs", currentFailureID, tss*100),
		)
	}

	return m
}

// WithExplicitHistory augments m with FPC and PHI computed from the supplied
// explicit history entries. m may be nil; a new Metrics is returned in that
// case. Returns m unchanged when fewer than minFPCEntries are supplied.
func WithExplicitHistory(m *model.Metrics, entries []model.MetricsHistoryEntry) *model.Metrics {
	if m == nil {
		m = &model.Metrics{}
	}

	n := len(entries)
	if n < minFPCEntries {
		return m
	}

	matchedCount := 0
	idCounts := map[string]int{}
	for _, e := range entries {
		if e.Matched {
			matchedCount++
			if e.FailureID != "" {
				idCounts[e.FailureID]++
			}
		}
	}

	fpc := round2(float64(matchedCount) / float64(n))
	m.FPC = &fpc

	if fpc < fpcWarnThreshold {
		m.DriftComponents = append(m.DriftComponents,
			fmt.Sprintf("low coverage: %d of %d analyzed runs had no playbook match", n-matchedCount, n),
		)
	}

	if n >= minPHIEntries {
		dominantShare := dominantFailureShare(idCounts, matchedCount)
		phi := round2(fpc * (1 - dominantShare))
		m.PHI = &phi

		if phi < phiWarnThreshold {
			top := topFailureID(idCounts)
			if top != "" {
				m.DriftComponents = append(m.DriftComponents,
					fmt.Sprintf("dominant pattern: %s is %.0f%% of matched failures", top, dominantShare*100),
				)
			}
		}
	}

	sort.Strings(m.DriftComponents)
	return m
}

// LoadHistoryFile reads a JSONL file where each line is a MetricsHistoryEntry.
// Lines that cannot be decoded are skipped. Returns an empty slice on any
// file-open error so callers can treat missing files as empty input.
func LoadHistoryFile(path string) ([]model.MetricsHistoryEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("open history file: %w", err)
	}
	defer f.Close()
	return decodeHistoryEntries(f)
}

func decodeHistoryEntries(r io.Reader) ([]model.MetricsHistoryEntry, error) {
	var entries []model.MetricsHistoryEntry
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		var e model.MetricsHistoryEntry
		if err := json.Unmarshal(line, &e); err != nil {
			// Skip malformed lines: external history files may have evolved.
			continue
		}
		entries = append(entries, e)
	}
	if err := sc.Err(); err != nil {
		return entries, fmt.Errorf("scan history file: %w", err)
	}
	return entries, nil
}

// dominantFailureShare returns the share of matched runs held by the single
// most common failure_id. Returns 0 when matchedCount is 0.
func dominantFailureShare(idCounts map[string]int, matchedCount int) float64 {
	if matchedCount == 0 {
		return 0
	}
	max := 0
	for _, c := range idCounts {
		if c > max {
			max = c
		}
	}
	return float64(max) / float64(matchedCount)
}

// topFailureID returns the most frequent failure ID in idCounts.
func topFailureID(idCounts map[string]int) string {
	best := ""
	bestCount := 0
	// Iterate in sorted key order for deterministic selection when counts tie.
	keys := make([]string, 0, len(idCounts))
	for k := range idCounts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if idCounts[k] > bestCount {
			bestCount = idCounts[k]
			best = k
		}
	}
	return best
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
