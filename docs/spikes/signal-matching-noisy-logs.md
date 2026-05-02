# Spike: Effective Methods for Matching Signals in Noisy Logs

**Status**: In Progress  
**Date**: 2026-04-28  
**Author**: research spike  
**Scope**: `internal/matcher`, `internal/engine/normalize.go`, playbook `match:` patterns

---

## Problem Statement

The current matcher achieves 87.2% coverage on the eval corpus using `strings.Contains` on
normalized log lines. The remaining ~13% of unmatched cases — plus the broader challenge of
keeping false-positive rates low as the playbook catalog grows — expose the limits of pure
substring matching in real-world CI log noise.

**What "noisy logs" means here:**

1. Timestamps, UUIDs, ANSI escapes, temp paths already stripped by `CanonicalizeLog`.
2. What stays: memory addresses, port numbers, line numbers, version strings, test function
   lists, stack-frame indices, locale-dependent messages, dynamic request IDs, duplicate lines.
3. Noise that defeats substring matching: morphological-variant error words ("timed out" vs
   "timeout" vs "exceeded deadline"), version-string debris in line bodies, high-frequency
   filler lines (progress bars, percentage counts) drowning signal lines.

---

## Current Implementation (Baseline)

### Normalization pipeline (`engine/normalize.go`)

| Stage | What it strips or replaces |
|-------|---------------------------|
| ANSI escapes | `\x1b[…m` sequences |
| RFC3339 / date-time / time-only | → `<timestamp>` |
| Temp paths (`/tmp`, `/var/folders`, `C:\Users\…\AppData\Local\Temp`) | → `<tmp>` |
| UUIDs | → `<id>` |
| Whitespace collapse + lowercase | by `NormalizeLine` / matcher `normalize()` |

**What it does NOT strip:** port numbers, memory addresses (`0x1a2b3c`), stack frame
ordinals (`#0 0x…`, `at line N`), version strings (`v1.2.3`), numeric counts, duplicate
identical lines.

### Matching logic (`internal/matcher/matcher.go`)

```
match.any  → OR semantics, IDF-weighted (1/N for patterns shared by N playbooks)
match.all  → AND semantics, flat +1.5 each, +2.0 compound bonus when all hit
match.partial → sub-cluster with configurable minimum threshold
match.none → hard veto (any match → score 0)
```

Confidence is post-hoc calibrated from score coverage + separation from second-best result.

### Known gaps from eval corpus

- Prior eval runs produced 10,881 unmatched clusters in generated gap-summary
  output; top clusters were:
  - Test-function-name list lines (pytest, unittest output with no error prefix)
  - Multilingual error messages (non-ASCII)
  - Numeric-only or address-only variants of known error strings
  - Short framework-specific assertion lines ("prop type `onClick` is invalid")

---

## Research Questions

1. **Q1**: Which normalization additions would reduce false-negative rates without increasing
   false-positives? (e.g., numeric normalization, hex-address stripping)
2. **Q2**: Can token-aware matching replace substring matching for a precision gain?
3. **Q3**: Does line-position weighting (end-of-log bias) improve matching on summary lines?
4. **Q4**: Can synonym/alias expansion in patterns cover morphological variants while staying
   deterministic?
5. **Q5**: Would proximity constraints (`all` patterns within N lines) improve precision on
   multi-signal playbooks?
6. **Q6**: Does frequency-within-log capping of evidence lines affect score calibration?

---

## Investigation Results

### 1. Normalization Improvements

#### 1a. Numeric normalization

**Rationale**: Log lines differ by specific numbers that are not semantically meaningful
for classification: port numbers, line numbers, memory addresses, exit codes sometimes,
repetition counts.

```
Before: "listen tcp 0.0.0.0:8080: bind: address already in use"
After:  "listen tcp 0.0.0.0:<port>: bind: address already in use"
```

**Useful patterns to collapse:**
- Hex addresses: `0x[0-9a-f]+` → `<addr>`
- Port in socket notation: `:\d{2,5}\b` in TCP/UDP contexts
- Memory size suffixes: `\d+(?:\.\d+)? (?:MB|GB|KB|MiB|GiB)` → `<size>`
- Exit codes (in specific contexts): "exited with N" → handled via pattern verbatim or
  wildcard

**Risk**: Over-normalizing numeric tokens removes distinguishing signal. Specific exit codes
(`124`, `137`) are intentional playbook patterns. Numeric normalization must be selective,
not global.

**Verdict**: Useful only for hex addresses and very long numeric tokens (>8 digits) that
are clearly dynamic IDs. Exit codes and small counts should remain verbatim.

#### 1b. Deduplication of repeated identical lines

**Rationale**: CI logs often contain hundreds of identical progress-bar or percentage lines.
These inflate line counts and can bias any coverage/frequency computation.

**Current state**: No deduplication. All lines evaluated independently.

**Proposed**: In `Analyze()`, deduplicate _consecutive_ identical normalized lines before
scoring, retaining one representative. Non-consecutive duplicates keep the first occurrence.

**Impact on scoring**: `match.any` already breaks on first hit per pattern per playbook
(inner loop has `break`), so deduplication mainly reduces CPU work and doesn't change
current score math. For future frequency-based methods it matters more.

**Verdict**: Safe and cheap to add as a pre-pass. Low priority for correctness, useful for
performance with large logs.

#### 1c. Stripping stack-frame prefixes

**Rationale**: Stack traces embed line numbers and function addresses that vary between runs:
```
at com.example.Foo.bar(Foo.java:42)
goroutine 1 [running]:
#0  0x00007f1234abcd in some_func (file.c:88)
```

The class/function name is stable; the line number and address are not. Current normalization
preserves `(Foo.java:42)` verbatim, which means patterns must omit the number (they do, via
substring). This is already handled implicitly: `strings.Contains("com.example.foo.bar", …)`
works fine as long as the pattern doesn't include the line number.

**Verdict**: Not needed. Current substring approach sidesteps this by pattern design.

---

### 2. Token-Aware Matching

#### 2a. Word-boundary matching vs. substring

**Rationale**: `strings.Contains("error", "errorcode")` is a false positive.
Current normalization collapses whitespace but does not add word-boundary guards.

**Example**:
- Pattern: `"error"`
- Line: `"errorcode: 0 (no error)"`
- Current match: YES (false positive)
- Token match: NO (correct)

**Implementation**:
```go
// Current
strings.Contains(line.Normalized, norm)

// Token-aware option A: whole-word boundary check
func containsWord(text, word string) bool {
    idx := strings.Index(text, word)
    for idx >= 0 {
        start := idx == 0 || !isWordChar(rune(text[idx-1]))
        end := idx+len(word) >= len(text) || !isWordChar(rune(text[idx+len(word)]))
        if start && end {
            return true
        }
        idx = strings.Index(text[idx+1:], word)
        if idx >= 0 { idx += ... } // adjusted offset
    }
    return false
}

// Token-aware option B: pre-tokenize lines into a set, match whole tokens
tokens := strings.Fields(line.Normalized)
tokenSet := map[string]struct{}{}
for _, t := range tokens { tokenSet[t] = struct{}{} }
_, ok := tokenSet[norm]
```

**Tradeoffs**:
- Option A: handles multi-word patterns correctly (phrase matching still works)
- Option B: single-token patterns only; breaks multi-word patterns (all current patterns
  are phrases of 1-4 words)
- Option A is preferred: it extends naturally, preserves phrase matching, and adds
  word-boundary precision

**Risk**: Some patterns are intentionally partial tokens (e.g., `"cannot"` to match
`"cannot find"`, `"cannot open"` etc.). Word-boundary matching preserves this because
`"cannot"` is itself a whole word in all those contexts.

**Verdict**: **High value.** Option A (phrase-level word-boundary check) reduces false
positives from generic substrings appearing inside longer tokens. Low implementation
risk. Fits determinism constraint.

#### 2b. Pre-tokenized line sets for O(1) lookup

**Current complexity**: O(P × L) per playbook, where P = patterns, L = lines. For a
catalog of 80 playbooks with 10 patterns each and 5,000 log lines: 4M comparisons per
analysis call.

**Pre-tokenized approach**: Build a `map[string][]int` (token → line numbers) once for the
full log, then pattern lookup is O(1) + O(hits) for evidence gathering.

```go
type LineIndex struct {
    tokenLines map[string][]int // normalized token → line indices
    lines      []model.Line
}

func BuildIndex(lines []model.Line) LineIndex { … }

func (idx LineIndex) LookupPhrase(phrase string) []int {
    // For single-word phrase, O(1).
    // For multi-word, anchor on rarest token then verify surroundings.
}
```

**Savings**: For large logs (>10k lines), index build is O(L) and lookup is O(1+hits).
This pays off at the per-analysis level when the same index is reused across all playbooks.

**Verdict**: **High value for performance.** Especially for large workflows. Requires a
new `LineIndex` struct passed to `matchPlaybook`; backward-compatible if opts are additive.

---

### 3. Synonym / Alias Expansion

**Rationale**: Many CI failure signals appear in morphologically different forms across
ecosystems:

| Concept | Observed variants |
|---------|------------------|
| Timeout | `timeout`, `timed out`, `deadline exceeded`, `context deadline exceeded`, `i/o timeout`, `read timeout`, `connection timed out` |
| Permission denied | `permission denied`, `access denied`, `forbidden`, `EACCES`, `operation not permitted` |
| Not found | `no such file or directory`, `not found`, `ENOENT`, `cannot find`, `does not exist`, `was not found` |
| Memory | `out of memory`, `OOM`, `cannot allocate memory`, `ENOMEM`, `heap exhausted` |

**Current workaround**: Each variant must be a separate pattern in `match.any`. This works
but creates long pattern lists that are hard to maintain.

**Proposed**: A shared `synonyms:` vocabulary at the playbook catalog level (or builtin to
the engine), so a pattern can reference a synonym group:

```yaml
# Proposed playbook syntax
match:
  any:
    - "@timeout"        # expands to all timeout variants
    - "@not-found"      # expands to all not-found variants
```

**Implementation complexity**: Medium. Requires:
1. A `synonyms.yaml` file (one-time authoring cost)
2. Pattern expansion at `Rank()` time or at catalog load
3. IDF weight computation on the expanded patterns (or on the group, not individual patterns)

**Determinism**: Fully deterministic — synonym expansion is static.

**Verdict**: **Medium value.** Useful for playbook authors; reduces pattern list length.
Does not address fundamentally new match types. Defer unless the catalog growth makes
pattern maintenance painful.

---

### 4. Line-Position Weighting

**Observation**: In CI logs, error summaries typically appear near the end:
- pytest: `FAILED tests/…` lines appear in a final block
- Go test: `--- FAIL: TestXxx` followed by failure summary
- Maven: `[ERROR] BUILD FAILURE` is the last meaningful line
- GitHub Actions: step failure summary appears after the log body

**Proposal**: Give higher weight to lines in the final 10-20% of the log, or to lines after
a detected "failure header" signal (e.g., `BUILD FAILURE`, `FAILED`, `Error:`).

```go
// Position weight: linear decay from end
func linePositionWeight(lineIdx, totalLines int) float64 {
    if totalLines == 0 { return 1.0 }
    tail := float64(totalLines) * 0.2
    distance := float64(totalLines - 1 - lineIdx)
    if distance < tail {
        return 1.0 + (tail-distance)/tail*0.5  // up to 1.5× for last 20%
    }
    return 1.0
}
```

**Tradeoffs**:
- Helps: summary-only playbooks that have one diagnostic line at EOF
- Hurts: playbooks where the signal appears early (e.g., compilation error on line 3 of a
  10,000-line build log)
- Breaks determinism guarantee? No — position is deterministic given same input.
- Breaks score stability? Yes — adding more lines to a log shifts all positions. This is
  **a correctness hazard**: the same matching line would score differently depending on how
  many unrelated lines follow it.

**Verdict**: **Low value, high risk for score instability.** Skip. Instead, consider
detecting "failure section" markers as a context signal and boosting stage/section score
separately (already partially done via `stageBonus`).

---

### 5. Proximity Constraints for `match.all`

**Current**: `match.all` patterns must ALL appear in the log, but with no constraint on
_where_ they appear. Two patterns could match lines 1 and 4,999 of a 5,000-line log and
still receive the compound bonus.

**Problem**: High false-positive risk for generic `all` patterns like `["error", "timeout"]`
that commonly co-occur in unrelated contexts within a long log.

**Proposal**: Add an optional `within_lines: N` constraint to `MatchSpec.All`:

```yaml
match:
  all:
    - "connection refused"
    - "retrying"
  within_lines: 10   # both must appear within 10 lines of each other
```

**Implementation**:
```go
// For match.all with proximity, build line-hit map then check adjacency:
func allPatternsProximate(patterns []string, lines []model.Line, within int) bool {
    hitLines := map[string]int{} // pattern → first matching line number
    for _, pat := range patterns {
        norm := normalize(pat)
        for _, line := range lines {
            if strings.Contains(line.Normalized, norm) {
                hitLines[norm] = line.Number
                break
            }
        }
    }
    if len(hitLines) != len(patterns) { return false }
    nums := make([]int, 0, len(hitLines))
    for _, n := range hitLines { nums = append(nums, n) }
    sort.Ints(nums)
    return nums[len(nums)-1]-nums[0] <= within
}
```

**Verdict**: **Medium value, medium complexity.** Useful for reducing false-positive
compound bonuses. Should be opt-in via playbook field (backward compatible). Most valuable
for high-specificity playbooks where co-occurrence in unrelated parts of a log is plausible.

---

### 6. Frequency Capping and Repeated-Line Handling

**Problem**: A log with 200 identical `"error: connection refused"` lines currently scores
the same as one with a single such line, because the inner loop breaks on first hit. But
if a future frequency-sensitive score is introduced (e.g., "signal strength proportional
to repetition count"), infinite-loop failures or retry storms would dominate.

**Current behavior is correct**: Break-on-first-hit means the current scorer is already
implicitly frequency-capped at 1 signal per pattern per playbook. No change needed.

**Future consideration**: If "repetition as severity amplifier" is ever wanted, the cap
should be explicit (e.g., max 3×) to prevent log flooding from distorting scores.

**Verdict**: No immediate action. Document the implicit cap.

---

### 7. Regex Pattern Support

**Current**: All patterns are pure strings matched via `strings.Contains`. Authors cannot
express "error code [0-9]+" without listing every variant.

**Proposal**: Allow patterns prefixed with `re:` to be treated as compiled regexes:

```yaml
match:
  any:
    - "re:exit(ed)? (with )?code [1-9][0-9]*"
    - "connection refused"    # plain substring, unchanged
```

**Tradeoffs**:
- Expressiveness: large gain for numeric-variant and optional-word patterns
- Safety: user-authored regex can be slow (catastrophic backtracking); must pre-validate
  and enforce a timeout or use RE2-style constraints
- IDF weighting: a regex pattern has weight 1.0 (can't be compared across playbooks by
  string identity); each regex is unique by definition
- Determinism: fully deterministic if patterns are static (they are)

**Implementation safety**:
```go
// Compile-on-load with RE2 (Go's regexp package is RE2-compliant, no backtracking risk)
import "regexp"
// Go's regexp is RE2-safe by default — no possessive or backreference operators.
// No extra timeout needed.
```

Go's `regexp` package uses RE2 semantics (no backtracking catastrophe), making this safe
to compile and run without timeouts.

**Verdict**: **High value, medium implementation cost.** RE2 patterns in Go are safe.
Opt-in via `re:` prefix preserves backward compatibility. Most impactful for numeric
variants that currently require many explicit patterns.

---

### 8. Assessment of "Noisy Line" Suppression

**Observation from eval gaps**: Many unmatched clusters consist of:
- Lines that are test function name lists (not error lines): `test_repository_package_order`
- Progress indicators: `Downloading… 45%`
- Boilerplate CI group headers: `##[group]Run tests`

These are not errors; the issue is that the correct playbook pattern simply isn't in the
catalog for the actual error. Suppressing these lines would not improve match rate — it
would just make the log shorter.

**Verdict**: Noisy-line suppression is the wrong frame for improving match rate. The gaps
are _catalog coverage_ problems (missing playbook patterns), not normalization problems.
Normalization improvements help with false-positives; catalog improvements help with
false-negatives.

---

## Ranking and Recommendations

| Method | Value | Risk | Effort | Priority |
|--------|-------|------|--------|----------|
| **RE2 regex pattern support** (`re:` prefix) | High | Low (RE2-safe) | M | **P1** |
| **Pre-tokenized line index** (O(1) lookup) | High (perf) | Low | M | **P1** |
| **Word-boundary phrase matching** | High (precision) | Low | S | **P2** |
| **`within_lines` proximity for `match.all`** | Medium | Low | M | **P2** |
| **Synonym/alias expansion** (`@group`) | Medium | Low | M | **P3** |
| **Hex-address normalization** | Low | Very low | S | **P3** |
| **Consecutive-line deduplication** | Low (CPU) | Very low | S | **P4** |
| **Line-position weighting** | Low | High (score instability) | M | **Do not implement** |

### Recommended immediate actions

1. **RE2 regex support** (`re:` prefix on patterns): Highest leverage for reducing
   pattern list sprawl on numeric-variant patterns. Implement as opt-in — no existing
   patterns break. Go `regexp` is RE2-safe.

2. **Word-boundary phrase matching**: Replace raw `strings.Contains` with a
   phrase-level word-boundary check. Drop-in replacement with lower false-positive rate
   for single-word patterns like `"error"`, `"failed"`, `"timeout"`.

3. **Pre-tokenized line index**: Extract into a `LineIndex` struct built once at the
   start of `Rank()`; each playbook's patterns do index lookups instead of full line
   scans. Transparent performance improvement, especially for large logs.

### What not to implement

- **Line position weighting**: Score values become dependent on log length, breaking
  the stability guarantee. Stage/section inference (already present) is the better proxy.
- **ML-based matching**: Excluded by explicit product rule (no LLM/ML in product logic).
- **Fuzzy edit-distance matching**: Non-deterministic in edge cases, hard to explain in
  evidence output, and unnecessary given RE2 regex covers the real use cases.

---

## Prototype / Testing Notes

### Test for word-boundary false-positive reduction

The current test suite in `matcher_test.go` does not have a case where a short pattern
accidentally matches inside a longer token. A regression test should be added:

```go
func TestRankWordBoundaryNoFalsePositive(t *testing.T) {
    pb := model.Playbook{
        ID:    "error-exact",
        Match: model.MatchSpec{Any: []string{"error"}},
    }
    lines := []model.Line{
        {Original: "errorcode: 0", Normalized: "errorcode: 0"},  // should NOT match
        {Original: "Exit error", Normalized: "exit error"},       // SHOULD match
    }
    // With current strings.Contains, both match.
    // With word-boundary check, only the second matches.
}
```

### Test for RE2 pattern matching

```go
func TestRankRegexPattern(t *testing.T) {
    pb := model.Playbook{
        ID:    "exit-nonzero",
        Match: model.MatchSpec{Any: []string{"re:exited? with (code )?[1-9][0-9]*"}},
    }
    lines := []model.Line{
        {Original: "Process exited with code 137", Normalized: "process exited with code 137"},
        {Original: "Process exited with 1", Normalized: "process exited with 1"},
        {Original: "Exit code: 0", Normalized: "exit code: 0"},  // should NOT match
    }
    results := Rank([]model.Playbook{pb}, lines, model.Context{})
    // Expect 1 result, 2 evidence lines
}
```

---

## External Resources

- [Go `regexp` package (RE2 semantics)](https://pkg.go.dev/regexp) — no catastrophic
  backtracking; safe to use without timeouts
- [RE2 syntax reference](https://github.com/google/re2/wiki/Syntax) — character classes,
  alternation, anchors
- [IDF weighting in IR](https://en.wikipedia.org/wiki/Tf%E2%80%93idf) — the current
  scheme implements a simplified IDF; BM25's saturation term is not needed at this scale
- [Aho-Corasick multi-pattern matching](https://en.wikipedia.org/wiki/Aho%E2%80%93Corasick_algorithm) —
  O(n+m) multi-pattern scan; relevant if pattern count grows past ~1,000 and CPU becomes
  a bottleneck
- Eval corpus gap analysis — generated under `eval-work/` by `make eval-run`;
  empirical source of unmatched cluster patterns when recreated locally

---

## Decision / Recommendation

**Primary recommendation**: Implement RE2 regex pattern support as an opt-in `re:` prefix.
This is the highest-leverage improvement: it unblocks numeric-variant patterns, exit-code
ranges, and optional-word variants without touching the existing substring path.

**Secondary recommendation**: Replace `strings.Contains` with a word-boundary phrase
check. This is a small code change with a meaningful precision improvement — eliminating
false positives from short patterns matching inside longer tokens.

**Tertiary recommendation**: Introduce a `LineIndex` for O(1) pattern lookup in
`RankPrecomputed`, reducing per-analysis cost from O(P×L) to O(L + P).

All three recommendations are:
- Deterministic
- Backward-compatible (opt-in or drop-in)
- Free of ML/LLM dependence
- Testable with the existing fixture corpus

---

## Status History

| Date | Status | Note |
|------|--------|------|
| 2026-04-28 | Complete | Initial research, all questions answered |
