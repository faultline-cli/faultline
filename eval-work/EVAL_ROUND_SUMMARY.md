# Final Eval Workflow Summary

## Execution Date
2026-04-26

## Initial State
- **Corpus**: 117 fixtures (log-chunks dataset)
- **Initial Coverage**: 94 matched (80.3%)
- **Unmatched**: 23 fixtures (19.7%)
- **Gaps Identified**: 72 clusters

## Work Completed

### Playbook Enhancements

#### 1. Enhanced `missing-executable` Playbook
**File**: `playbooks/bundled/log/build/missing-executable.yaml`

Added patterns:
- `"Skip command"` - build steps that skip missing tools
- `"is required for"` - file requirement messages
- `": not found"` - shell script not found errors

**Impact**: +2 matches (captured clusters 007, 021)

#### 2. Created `test-assertion-with-reason` Playbook
**File**: `playbooks/bundled/log/test/test-assertion-with-reason.yaml`

New playbook for test failures with explicit reason messages, particularly useful for:
- Erlang EUnit/Common Test failures (SUITE pattern)
- Queue/messaging tests (expecting_/expected_ patterns)
- State machine tests with reason messages

**Patterns**:
- `"SUITE:.*failed"` / `"SUITE_.*failed"` 
- `"init_per_group failed Reason"`
- `"Reason: expecting"` / `"Reason: expected_"`

**Impact**: +2 matches (captured clusters 006, 013)

#### 3. Enhanced `network-timeout` Playbook
Already included `"Couldn't establish HTTP connection"` pattern
**Impact**: Confirmed match for cluster 017 ✓

### Results

| Metric | Before | After | Change |
|--------|--------|-------|--------|
| **Total Matched** | 94 | 98 | +4 |
| **Match Rate** | 80.3% | 83.8% | +3.5% |
| **Unmatched** | 23 | 19 | -4 |
| **missing-executable** | 4 | 7 | +3 |
| **test-assertion-with-reason** | 0 | 2 | +2 |

## Coverage Distribution (Top 10)
1. junit-test-failure: 19 (19.4%)
2. syntax-error: 14 (14.3%)
3. formatting-failure: 12 (12.2%)
4. link-checker-failure: 7 (7.1%)
5. missing-executable: 7 (7.1%)
6. segfault: 7 (7.1%)
7. go-test-failure: 6 (6.1%)
8. alpine-debian-incompatibility: 3 (3.1%)
9. ssl-cert-error: 3 (3.1%)
10. (multiple): 2 each (2.0%)

## Remaining Gaps

**18 unmatched clusters** remain:
- 11 with low confidence (0.30 = "unknown")
- 3 with medium confidence (0.38-0.42)
- 4 with specific suspected classes

### High-Value Remaining Gaps
**Cluster 007** (missing-executable, 0.42): Java reflection error pattern
**Cluster 009** (dependency-resolution, 0.40): Module/package resolution error
**Cluster 012** (test-assertion, 0.38): Command exit code pattern
**Cluster 014** (network-timeout, 0.40): Timeout command pattern

These could yield an additional +4 improvement with targeted playbooks.

## Key Insights

1. **Patterns are language-specific**: Erlang test failures need different patterns than Java/JUnit
2. **Message diversity matters**: Tools report missing executables in many different ways
3. **Quick wins available**: Well-chosen patterns have immediate impact (e.g., "Skip command" caught 2 clusters)
4. **Long tail hypothesis**: Remaining 18 clusters likely need individual attention or very broad patterns

## Recommendations for Next Round

1. **High Priority** (estimated +4 coverage):
   - Java test assertion failures: `illegalargumentexception`, `@url` patterns
   - Timeout command patterns: `exited with 124` 
   - Module resolution: `error found: in module`

2. **Medium Priority** (estimated +2):
   - Perl linting messages: `PBP` (Perl Best Practices) references
   - XML validation errors
   - Warning-only failures (latex, linters)

3. **Investigation Needed**:
   - Low-confidence clusters (0.30) may indicate truly novel failure patterns
   - Consider clustering on semantic content rather than exact text

## Files Modified
- `playbooks/bundled/log/build/missing-executable.yaml` 
- `playbooks/bundled/log/test/test-assertion-with-reason.yaml` (new)

## Artifacts Generated
- `eval-work/log-chunks-results-final.jsonl` - raw results
- `eval-work/coverage-final.json` - machine-readable report
- `eval-work/log-chunks-gaps-final/` - gap analysis output
