// Package matcher implements deterministic pattern matching and scoring of
// playbooks against normalised log lines.
package matcher

import (
	"math"
	"regexp"
	"sort"
	"strings"

	"faultline/internal/model"
)

// evidenceItemMax is the maximum number of unique evidence lines collected
// per playbook match. Evidence beyond this limit adds no diagnostic value
// and would bloat JSON output and the SQLite store.
const evidenceItemMax = 20

// regexPrefix is the pattern prefix that signals a RE2 regex rather than a plain substring.
const regexPrefix = "re:"

// patternKey returns a stable map key for a pattern.
// For plain patterns it returns the normalized (lowercase, whitespace-collapsed) form.
// For re:-prefixed patterns it returns "re:" + the lowercased regex source.
// Returns "" when the pattern is empty after processing and should be skipped.
func patternKey(pat string) string {
	if strings.HasPrefix(pat, regexPrefix) {
		src := strings.TrimSpace(strings.TrimPrefix(pat, regexPrefix))
		if src == "" {
			return ""
		}
		return regexPrefix + strings.ToLower(src)
	}
	return normalize(pat)
}

// Rank matches every playbook against lines and returns results sorted by
// score descending, then confidence descending, then playbook ID ascending.
// Only playbooks with at least one matching pattern are included.
//
// match.any patterns are weighted by inverse document frequency: a pattern
// shared by N playbooks contributes 1/N to the score, so generic terms that
// appear across many playbooks are automatically less decisive than patterns
// unique to a single playbook. match.all and match.none semantics are unchanged.
func Rank(playbooks []model.Playbook, lines []model.Line, ctx model.Context) []model.Result {
	return RankPrecomputed(playbooks, computeAnyWeights(playbooks), lines, ctx)
}

// RankPrecomputed is identical to Rank but accepts pre-computed IDF weights so
// callers that issue many analyses against the same playbook set can compute
// the weights once and reuse them across calls.
func RankPrecomputed(playbooks []model.Playbook, weights map[string]float64, lines []model.Line, ctx model.Context) []model.Result {
	idx := buildLineIndex(playbooks, lines)
	results := make([]model.Result, 0, len(playbooks))
	for _, pb := range playbooks {
		r := matchPlaybook(pb, idx, ctx, weights)
		if r.Score == 0 {
			continue
		}
		results = append(results, r)
	}

	sort.Slice(results, func(i, j int) bool {
		a, b := results[i], results[j]
		if a.Score != b.Score {
			return a.Score > b.Score
		}
		return a.Playbook.ID < b.Playbook.ID
	})
	calibrateConfidence(results)

	sort.Slice(results, func(i, j int) bool {
		a, b := results[i], results[j]
		if a.Score != b.Score {
			return a.Score > b.Score
		}
		if a.Confidence != b.Confidence {
			return a.Confidence > b.Confidence
		}
		if len(a.Evidence) != len(b.Evidence) {
			return len(a.Evidence) > len(b.Evidence)
		}
		return a.Playbook.ID < b.Playbook.ID
	})

	return results
}

// matchPlaybook scores a single playbook against lines using the following
// rules:
//
//   - Each matched any-pattern  → +weight (1/N where N = playbooks sharing that pattern)
//   - Each matched all-pattern  → +1.5 (flat; AND semantics already discriminate)
//   - Each matched partial-group pattern → +0.75
//   - Partial group threshold met → +1.0 bonus
//   - All all-patterns matched  → +2.0 bonus (gated by within_lines when set)
//   - Stage hint matches ctx    → +0.75
//   - Playbook base_score       → added unconditionally when patterns match
//
// Confidence is calibrated from the matched score and competitive separation
// after the full ranked result set is known.
func matchPlaybook(pb model.Playbook, idx lineIndex, ctx model.Context, weights map[string]float64) model.Result {
	evidence := make([]string, 0)
	seenEvidence := make(map[string]struct{})

	addEvidence := func(line string) {
		if len(evidence) >= evidenceItemMax {
			return
		}
		if _, ok := seenEvidence[line]; !ok {
			evidence = append(evidence, line)
			seenEvidence[line] = struct{}{}
		}
	}

	// Score any-patterns (OR semantics, IDF-weighted).
	//
	// For playbooks that extend a parent, anyScore is computed from the
	// child's own native patterns only (NativeAny), not from the full
	// merged Match.Any that includes inherited parent patterns. This
	// prevents a child from tying with its parent on logs where only
	// generic parent patterns fire, preserving the parent's confidence.
	// Inherited patterns from Match.Any are still collected as evidence
	// so context is preserved in the output even when they do not score.
	scoringPatterns := pb.Match.Any
	if pb.Extends != "" && pb.NativeAny != nil {
		scoringPatterns = pb.NativeAny
	}

	anyScore := 0.0
	for _, pat := range scoringPatterns {
		k := patternKey(pat)
		if k == "" {
			continue
		}
		w := weights[k]
		if w == 0 {
			w = 1.0
		}
		if orig, ok := idx.firstOriginal(k); ok {
			anyScore += w
			addEvidence(orig)
		}
	}
	// Collect evidence from inherited patterns without contributing to score.
	if pb.Extends != "" {
		for _, pat := range pb.Match.Any {
			k := patternKey(pat)
			if k == "" {
				continue
			}
			if orig, ok := idx.firstOriginal(k); ok {
				addEvidence(orig)
			}
		}
	}

	// Score all-patterns (AND semantics; partial matches still accumulate).
	allHits := 0
	allComplete := len(pb.Match.All) > 0
	allKeys := make([]string, 0, len(pb.Match.All))
	for _, pat := range pb.Match.All {
		k := patternKey(pat)
		if k == "" {
			continue
		}
		allKeys = append(allKeys, k)
		if orig, ok := idx.firstOriginal(k); ok {
			allHits++
			addEvidence(orig)
		} else {
			allComplete = false
		}
	}

	// When within_lines is set, the compound bonus is only awarded when all
	// match.all patterns appear within that many lines of each other.
	if allComplete && pb.Match.WithinLines > 0 {
		allComplete = allPatternsWithinWindow(allKeys, idx, pb.Match.WithinLines)
	}

	partialScore := 0.0
	partialHits := 0
	for _, group := range pb.Match.Partial {
		hits := 0
		for _, pat := range group.Patterns {
			k := patternKey(pat)
			if k == "" {
				continue
			}
			if orig, ok := idx.firstOriginal(k); ok {
				hits++
				addEvidence(orig)
			}
		}
		if hits == 0 {
			continue
		}
		partialHits += hits
		partialScore += float64(hits) * 0.75
		if hits >= group.Minimum {
			partialScore += 1.0
		}
	}

	for _, pat := range pb.Match.None {
		k := patternKey(pat)
		if k == "" {
			continue
		}
		if idx.hasMatch(k) {
			return model.Result{}
		}
	}

	if anyScore == 0 && allHits == 0 && partialHits == 0 {
		return model.Result{} // no match
	}

	score := pb.BaseScore + anyScore + float64(allHits)*1.5 + partialScore
	if allComplete {
		score += compoundBonus(allComplete)
	}

	// Stage bonus (does not contribute to confidence calculation).
	score += stageBonus(pb, ctx)

	return model.Result{
		Playbook:   pb,
		Detector:   "log",
		Score:      math.Round(score*100) / 100,
		Confidence: 0,
		Evidence:   evidence,
		EvidenceBy: model.EvidenceBundle{
			Triggers: buildLogEvidence(evidence),
		},
		Explanation: model.ResultExplanation{
			TriggeredBy: evidence,
		},
		Breakdown: model.ScoreBreakdown{
			BaseSignalScore:     math.Round((pb.BaseScore+anyScore+float64(allHits)*1.5+partialScore)*100) / 100,
			CompoundSignalBonus: math.Round(compoundBonus(allComplete)*100) / 100,
			HotPathMultiplier:   math.Round(stageBonus(pb, ctx)*100) / 100,
			FinalScore:          math.Round(score*100) / 100,
		},
	}
}

// lineIndex precomputes which lines match each distinct pattern across all playbooks.
// Building it once per Rank call amortises the O(lines) scan cost across all playbooks
// that share a pattern. Regex patterns (re: prefix) are compiled once and reused.
type lineIndex struct {
	hits  map[string][]int // patternKey → indices into lines where the pattern matched
	lines []model.Line
}

// buildLineIndex scans every distinct pattern from every playbook against lines once,
// producing a lineIndex that matchPlaybook can query in O(1) per pattern.
func buildLineIndex(playbooks []model.Playbook, lines []model.Line) lineIndex {
	type patEntry struct {
		re  *regexp.Regexp // non-nil for regex patterns
		sub string         // non-empty for plain substring patterns
	}
	pats := make(map[string]patEntry)

	add := func(pat string) {
		k := patternKey(pat)
		if k == "" {
			return
		}
		if _, exists := pats[k]; exists {
			return
		}
		if strings.HasPrefix(k, regexPrefix) {
			src := strings.TrimPrefix(k, regexPrefix)
			re, err := regexp.Compile(src)
			if err != nil {
				// Invalid regex: omit from index so it never matches.
				return
			}
			pats[k] = patEntry{re: re}
		} else {
			pats[k] = patEntry{sub: k}
		}
	}

	for _, pb := range playbooks {
		for _, p := range pb.Match.Any {
			add(p)
		}
		for _, p := range pb.Match.All {
			add(p)
		}
		for _, p := range pb.Match.None {
			add(p)
		}
		for _, g := range pb.Match.Partial {
			for _, p := range g.Patterns {
				add(p)
			}
		}
	}

	hits := make(map[string][]int, len(pats))
	for k, pe := range pats {
		for i, line := range lines {
			var matched bool
			if pe.re != nil {
				matched = pe.re.MatchString(line.Normalized)
			} else {
				matched = containsPhrase(line.Normalized, pe.sub)
			}
			if matched {
				hits[k] = append(hits[k], i)
			}
		}
	}
	return lineIndex{hits: hits, lines: lines}
}

// firstOriginal returns the Original text of the first line matched by key, and true.
// Returns ("", false) when no line matches.
func (idx lineIndex) firstOriginal(key string) (string, bool) {
	if idxs := idx.hits[key]; len(idxs) > 0 {
		return idx.lines[idxs[0]].Original, true
	}
	return "", false
}

// hasMatch reports whether any line in the index matches the given pattern key.
func (idx lineIndex) hasMatch(key string) bool {
	return len(idx.hits[key]) > 0
}

// isWordByte reports whether the ASCII byte c is a word character
// (lowercase letter, digit, or underscore).  Normalised lines are already
// lowercased, so only the lowercase range needs to be tested.
func isWordByte(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_'
}

// containsPhrase reports whether phrase appears in text at a word boundary.
//
// Boundary rules:
//   - Start boundary: if the phrase begins with a word byte, the character
//     immediately before the match must not be a word byte (or the match is
//     at position 0).  This prevents "error" from matching inside "preerror".
//   - End boundary: only enforced for "simple" patterns — those composed
//     entirely of word bytes (a-z, 0-9, _) with no spaces or punctuation.
//     For a simple pattern, the character immediately after the match must
//     not be a word byte.  This prevents "concurrent" from matching inside
//     "concurrently" or "error" from matching inside "errorcode".
//
// Multi-word or mixed patterns (e.g. "environment variable", "exit code 1",
// "error[E") intentionally serve as sub-string probes — "environment variable"
// must match "environment variables", "exit code 1" must match "exit code 128",
// and "error[E" must match "error[E0502]".  For those patterns only the start
// boundary is checked.
func containsPhrase(text, phrase string) bool {
	n := len(phrase)
	if n == 0 {
		return false
	}
	checkStart := isWordByte(phrase[0])

	// A pattern is "simple" only when every byte is a word byte.  Simple
	// patterns get end-boundary enforcement; all others do not.
	isSimple := checkStart
	if isSimple {
		for _, c := range []byte(phrase) {
			if !isWordByte(c) {
				isSimple = false
				break
			}
		}
	}
	checkEnd := isSimple && isWordByte(phrase[n-1])

	if !checkStart && !checkEnd {
		return strings.Contains(text, phrase)
	}
	start := 0
	for {
		i := strings.Index(text[start:], phrase)
		if i < 0 {
			return false
		}
		i += start
		startOK := !checkStart || i == 0 || !isWordByte(text[i-1])
		endOK := !checkEnd || i+n == len(text) || !isWordByte(text[i+n])
		if startOK && endOK {
			return true
		}
		start = i + 1
		if start >= len(text) {
			return false
		}
	}
}

// allPatternsWithinWindow reports whether every key in keys has at least one
// matching line such that all such lines fall within a window of `within`
// lines (measured by position in the original lines slice).  It uses a
// minimal-span sliding window over all hits sorted by position.
func allPatternsWithinWindow(keys []string, idx lineIndex, within int) bool {
	type entry struct{ pos, pat int }
	var hits []entry
	for pi, k := range keys {
		for _, pos := range idx.hits[k] {
			hits = append(hits, entry{pos: pos, pat: pi})
		}
	}
	if len(hits) == 0 {
		return false
	}
	sort.Slice(hits, func(i, j int) bool { return hits[i].pos < hits[j].pos })

	total := len(keys)
	inWindow := make(map[int]int, total)
	covered := 0
	left := 0
	for right := 0; right < len(hits); right++ {
		h := hits[right]
		if inWindow[h.pat] == 0 {
			covered++
		}
		inWindow[h.pat]++
		for covered == total {
			if hits[right].pos-hits[left].pos <= within {
				return true
			}
			lh := hits[left]
			inWindow[lh.pat]--
			if inWindow[lh.pat] == 0 {
				covered--
			}
			left++
		}
	}
	return false
}

func calibrateConfidence(results []model.Result) {
	if len(results) == 0 {
		return
	}

	topScore := results[0].Score
	topScoreCount := 0
	secondScore := 0.0
	for _, result := range results {
		if result.Score == topScore {
			topScoreCount++
			continue
		}
		secondScore = result.Score
		break
	}

	for i := range results {
		competitorScore := topScore
		if results[i].Score == topScore && topScoreCount == 1 {
			competitorScore = secondScore
		}
		results[i].Confidence = confidenceFromScores(results[i].Score, competitorScore)
	}
}

func confidenceFromScores(score, competitorScore float64) float64 {
	if score <= 0 {
		return 0
	}

	coverage := score / (score + 1)
	separation := 0.0
	if competitorScore < score {
		separation = 1 - (competitorScore / score)
	}

	return math.Round(((coverage+separation)/2)*100) / 100
}

func buildLogEvidence(lines []string) []model.Evidence {
	out := make([]model.Evidence, 0, len(lines))
	for _, line := range lines {
		out = append(out, model.Evidence{
			Kind:   model.EvidenceTrigger,
			Label:  "Matched log evidence",
			Detail: line,
			Source: "log",
		})
	}
	return out
}

func compoundBonus(allComplete bool) float64 {
	if allComplete {
		return 2.0
	}
	return 0
}

func stageBonus(pb model.Playbook, ctx model.Context) float64 {
	if ctx.Stage == "" {
		return 0
	}
	for _, hint := range pb.StageHints {
		if strings.EqualFold(hint, ctx.Stage) {
			return 0.75
		}
	}
	return 0
}

// AnyWeights returns the IDF weight map for match.any patterns across the
// given playbook set. Pass the result to RankPrecomputed when analysing
// multiple log sources against the same playbook set.
func AnyWeights(playbooks []model.Playbook) map[string]float64 {
	return computeAnyWeights(playbooks)
}

// computeAnyWeights returns a map from normalized pattern string to its IDF
// weight: 1.0 / count, where count is the number of distinct playbooks that
// list that pattern in match.any. Patterns unique to one playbook keep
// weight 1.0; patterns shared by N playbooks each contribute 1/N.
//
// For playbooks that extend a parent (NativeAny is set), only the native
// patterns are counted. Inherited patterns do not affect IDF weights since
// they are not scored in the child (see matchPlaybook). This preserves the
// IDF weight of established base playbooks when specialised children extend
// them, so inheritance does not dilute detection confidence.
func computeAnyWeights(playbooks []model.Playbook) map[string]float64 {
	counts := make(map[string]int, len(playbooks)*8)
	for _, pb := range playbooks {
		// For child playbooks, use only their native (non-inherited) patterns
		// for IDF counting, matching the scoring semantics in matchPlaybook.
		pats := pb.Match.Any
		if pb.Extends != "" && pb.NativeAny != nil {
			pats = pb.NativeAny
		}
		seen := make(map[string]struct{}, len(pats))
		for _, p := range pats {
			k := patternKey(p)
			if k == "" {
				continue
			}
			if _, ok := seen[k]; ok {
				continue
			}
			seen[k] = struct{}{}
			counts[k]++
		}
	}
	weights := make(map[string]float64, len(counts))
	for p, c := range counts {
		weights[p] = 1.0 / float64(c)
	}
	return weights
}

// normalize lower-cases s and collapses internal whitespace.
func normalize(s string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(s)), " "))
}
