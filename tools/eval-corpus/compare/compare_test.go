package compare_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"faultline/tools/eval-corpus/compare"
	"faultline/tools/eval-corpus/model"
	"faultline/tools/eval-corpus/report"
)

func makeResult(id string, matched bool, failureID string) model.EvalResult {
	return model.EvalResult{
		FixtureID: id,
		Matched:   matched,
		FailureID: failureID,
	}
}

func TestCompareMatchRates(t *testing.T) {
	baseline := []model.EvalResult{
		makeResult("a", true, "auth"),
		makeResult("b", true, "auth"),
		makeResult("c", false, ""),
		makeResult("d", false, ""),
	}
	current := []model.EvalResult{
		makeResult("a", true, "auth"),
		makeResult("b", true, "auth"),
		makeResult("c", true, "missing-exec"),
		makeResult("d", false, ""),
	}
	res := compare.Compare(baseline, current, compare.Options{})
	if res.BaselineMatchRate != 0.5 {
		t.Errorf("BaselineMatchRate = %.2f, want 0.50", res.BaselineMatchRate)
	}
	if res.CurrentMatchRate != 0.75 {
		t.Errorf("CurrentMatchRate = %.2f, want 0.75", res.CurrentMatchRate)
	}
	if res.Delta != 0.25 {
		t.Errorf("Delta = %.2f, want 0.25", res.Delta)
	}
}

func TestComparePassWhenAboveMinRate(t *testing.T) {
	results := []model.EvalResult{
		makeResult("a", true, "auth"),
		makeResult("b", true, "auth"),
		makeResult("c", false, ""),
	}
	res := compare.Compare(results, results, compare.Options{MinMatchRate: 0.50})
	if !res.Pass {
		t.Errorf("expected pass when match rate (0.67) >= min (0.50)")
	}
}

func TestCompareFailsWhenBelowMinRate(t *testing.T) {
	baseline := []model.EvalResult{makeResult("a", true, "x"), makeResult("b", false, "")}
	current := []model.EvalResult{makeResult("a", false, ""), makeResult("b", false, "")}
	res := compare.Compare(baseline, current, compare.Options{MinMatchRate: 0.72})
	if res.Pass {
		t.Error("expected fail when match rate (0.00) < min (0.72)")
	}
	if len(res.FailReasons) == 0 {
		t.Error("expected at least one fail reason")
	}
}

func TestCompareFailsOnCoverageDrop(t *testing.T) {
	baseline := []model.EvalResult{
		makeResult("a", true, "auth"),
		makeResult("b", true, "auth"),
	}
	current := []model.EvalResult{
		makeResult("a", true, "auth"),
		makeResult("b", false, ""),
	}
	res := compare.Compare(baseline, current, compare.Options{FailOnCoverageDrop: true})
	if res.Pass {
		t.Error("expected fail when coverage dropped")
	}
}

func TestCompareNoDropDoesNotFail(t *testing.T) {
	baseline := []model.EvalResult{makeResult("a", false, "")}
	current := []model.EvalResult{makeResult("a", true, "auth")}
	res := compare.Compare(baseline, current, compare.Options{FailOnCoverageDrop: true})
	if !res.Pass {
		t.Errorf("expected pass when coverage improved: %v", res.FailReasons)
	}
}

func TestComparePlaybookDeltas(t *testing.T) {
	baseline := []model.EvalResult{
		makeResult("a", true, "auth"),
		makeResult("b", true, "auth"),
		makeResult("c", true, "missing-exec"),
	}
	current := []model.EvalResult{
		makeResult("a", true, "auth"),
		makeResult("b", false, ""),
		makeResult("c", true, "missing-exec"),
	}
	res := compare.Compare(baseline, current, compare.Options{})
	if len(res.PlaybooksLostMatches) != 1 || res.PlaybooksLostMatches[0].FailureID != "auth" {
		t.Errorf("expected auth in lost matches, got %v", res.PlaybooksLostMatches)
	}
	if len(res.PlaybooksGainedMatches) != 0 {
		t.Errorf("expected no gained matches, got %v", res.PlaybooksGainedMatches)
	}
}

func TestCompareNewUnmatchedTags(t *testing.T) {
	baseline := []model.EvalResult{
		{FixtureID: "a", Matched: false, FirstLineTag: "tag1"},
	}
	current := []model.EvalResult{
		{FixtureID: "a", Matched: false, FirstLineTag: "tag1"},
		{FixtureID: "b", Matched: false, FirstLineTag: "tag2"}, // new tag
	}
	res := compare.Compare(baseline, current, compare.Options{})
	if len(res.NewUnmatchedTags) != 1 || res.NewUnmatchedTags[0] != "tag2" {
		t.Errorf("NewUnmatchedTags = %v, want [tag2]", res.NewUnmatchedTags)
	}
}

func TestCompareEmptyInputsPass(t *testing.T) {
	res := compare.Compare(nil, nil, compare.Options{MinMatchRate: 0.50})
	// Zero results means 0% match rate which is < 0.50 — gate should fail.
	// But there are no results at all, so the gate still applies.
	if res.Pass {
		t.Error("expected fail: 0 results → 0% match rate < min 0.50")
	}
}

func TestCompareDeterministicOutput(t *testing.T) {
	// Running the same compare twice should produce identical results.
	baseline := []model.EvalResult{makeResult("a", true, "auth"), makeResult("b", false, "")}
	current := []model.EvalResult{makeResult("a", false, ""), makeResult("b", false, "")}
	r1 := compare.Compare(baseline, current, compare.Options{})
	r2 := compare.Compare(baseline, current, compare.Options{})
	if r1.Pass != r2.Pass {
		t.Error("Pass differs between runs")
	}
	if r1.Delta != r2.Delta {
		t.Error("Delta differs between runs")
	}
}

// --- AttachNondeterminism ---

func TestAttachNondeterminismDoesNothingForNoDifferences(t *testing.T) {
	det := report.DeterminismResult{Deterministic: true}
	res := &compare.Result{Pass: true}
	compare.AttachNondeterminism(res, det, compare.Options{})
	if len(res.NondeterministicFixtureIDs) != 0 {
		t.Errorf("expected no non-deterministic IDs, got %v", res.NondeterministicFixtureIDs)
	}
	if !res.Pass {
		t.Error("expected Pass to remain true")
	}
}

func TestAttachNondeterminismPopulatesIDs(t *testing.T) {
	det := report.DeterminismResult{
		Deterministic: false,
		Differences: []report.Difference{
			{FixtureID: "fix-aaa", Field: "matched"},
			{FixtureID: "fix-bbb", Field: "failure_id"},
			{FixtureID: "fix-aaa", Field: "confidence"}, // duplicate — should be deduplicated
		},
	}
	res := &compare.Result{Pass: true}
	compare.AttachNondeterminism(res, det, compare.Options{})
	if len(res.NondeterministicFixtureIDs) != 2 {
		t.Errorf("expected 2 unique IDs, got %v", res.NondeterministicFixtureIDs)
	}
}

func TestAttachNondeterminismSkipsFixtureIDField(t *testing.T) {
	det := report.DeterminismResult{
		Deterministic: false,
		Differences: []report.Difference{
			{FixtureID: "fix-aaa", Field: "fixture_id"}, // should be skipped
		},
	}
	res := &compare.Result{Pass: true}
	compare.AttachNondeterminism(res, det, compare.Options{})
	if len(res.NondeterministicFixtureIDs) != 0 {
		t.Errorf("expected 0 IDs (fixture_id field skipped), got %v", res.NondeterministicFixtureIDs)
	}
}

func TestAttachNondeterminismFailsGateWhenEnabled(t *testing.T) {
	det := report.DeterminismResult{
		Deterministic: false,
		Differences: []report.Difference{
			{FixtureID: "fix-aaa", Field: "matched"},
		},
	}
	res := &compare.Result{Pass: true}
	compare.AttachNondeterminism(res, det, compare.Options{FailOnNewNondeterminism: true})
	if res.Pass {
		t.Error("expected Pass=false when FailOnNewNondeterminism is set")
	}
	if len(res.FailReasons) == 0 {
		t.Error("expected at least one fail reason")
	}
}

func TestAttachNondeterminismDoesNotFailGateWhenDisabled(t *testing.T) {
	det := report.DeterminismResult{
		Deterministic: false,
		Differences: []report.Difference{
			{FixtureID: "fix-aaa", Field: "matched"},
		},
	}
	res := &compare.Result{Pass: true}
	compare.AttachNondeterminism(res, det, compare.Options{FailOnNewNondeterminism: false})
	if !res.Pass {
		t.Error("expected Pass to remain true when FailOnNewNondeterminism is disabled")
	}
}

// --- PrintTextReport ---

func TestPrintTextReportPass(t *testing.T) {
	res := compare.Result{
		Pass:              true,
		GeneratedAt:       "2026-04-26T00:00:00Z",
		BaselineTotal:     10,
		CurrentTotal:      10,
		BaselineMatchRate: 0.70,
		CurrentMatchRate:  0.80,
		Delta:             0.10,
	}
	var buf bytes.Buffer
	compare.PrintTextReport(&buf, res)
	out := buf.String()
	if !strings.Contains(out, "Gate: PASS") {
		t.Errorf("expected 'Gate: PASS' in output:\n%s", out)
	}
	if !strings.Contains(out, "70.00%") {
		t.Errorf("expected baseline rate in output:\n%s", out)
	}
}

func TestPrintTextReportFail(t *testing.T) {
	res := compare.Result{
		Pass:              false,
		GeneratedAt:       "2026-04-26T00:00:00Z",
		BaselineTotal:     10,
		CurrentTotal:      10,
		BaselineMatchRate: 0.80,
		CurrentMatchRate:  0.60,
		Delta:             -0.20,
		FailReasons:       []string{"coverage dropped by 0.2000"},
	}
	var buf bytes.Buffer
	compare.PrintTextReport(&buf, res)
	out := buf.String()
	if !strings.Contains(out, "Gate: FAIL") {
		t.Errorf("expected 'Gate: FAIL' in output:\n%s", out)
	}
	if !strings.Contains(out, "coverage dropped") {
		t.Errorf("expected fail reason in output:\n%s", out)
	}
}

func TestPrintTextReportIncludesLostAndGainedPlaybooks(t *testing.T) {
	res := compare.Result{
		Pass:              false,
		BaselineMatchRate: 0.8,
		CurrentMatchRate:  0.7,
		Delta:             -0.1,
		PlaybooksLostMatches: []compare.PlaybookDelta{
			{FailureID: "docker-auth", BaselineCount: 5, CurrentCount: 3, Delta: -2},
		},
		PlaybooksGainedMatches: []compare.PlaybookDelta{
			{FailureID: "missing-exec", BaselineCount: 1, CurrentCount: 3, Delta: 2},
		},
		NondeterministicFixtureIDs: []string{"fix-abc"},
	}
	var buf bytes.Buffer
	compare.PrintTextReport(&buf, res)
	out := buf.String()
	if !strings.Contains(out, "docker-auth") {
		t.Errorf("expected docker-auth in lost playbooks:\n%s", out)
	}
	if !strings.Contains(out, "missing-exec") {
		t.Errorf("expected missing-exec in gained playbooks:\n%s", out)
	}
	if !strings.Contains(out, "fix-abc") {
		t.Errorf("expected non-deterministic fixture in output:\n%s", out)
	}
}

// --- PrintMarkdownReport ---

func TestPrintMarkdownReportPass(t *testing.T) {
	res := compare.Result{
		Pass:              true,
		GeneratedAt:       "2026-04-26T00:00:00Z",
		BaselineTotal:     10,
		CurrentTotal:      12,
		BaselineMatchRate: 0.70,
		CurrentMatchRate:  0.75,
		Delta:             0.05,
	}
	var buf bytes.Buffer
	compare.PrintMarkdownReport(&buf, res)
	out := buf.String()
	if !strings.Contains(out, "✅ PASS") {
		t.Errorf("expected PASS status in markdown output:\n%s", out)
	}
	if !strings.Contains(out, "## Coverage") {
		t.Errorf("expected Coverage section:\n%s", out)
	}
}

func TestPrintMarkdownReportFail(t *testing.T) {
	res := compare.Result{
		Pass:              false,
		GeneratedAt:       "2026-04-26T00:00:00Z",
		BaselineTotal:     10,
		CurrentTotal:      10,
		BaselineMatchRate: 0.80,
		CurrentMatchRate:  0.60,
		Delta:             -0.20,
		FailReasons:       []string{"coverage dropped by 0.2000"},
		PlaybooksLostMatches: []compare.PlaybookDelta{
			{FailureID: "docker-auth", BaselineCount: 4, CurrentCount: 2, Delta: -2},
		},
		PlaybooksGainedMatches: []compare.PlaybookDelta{
			{FailureID: "new-failure", BaselineCount: 0, CurrentCount: 2, Delta: 2},
		},
	}
	var buf bytes.Buffer
	compare.PrintMarkdownReport(&buf, res)
	out := buf.String()
	if !strings.Contains(out, "❌ FAIL") {
		t.Errorf("expected FAIL status:\n%s", out)
	}
	if !strings.Contains(out, "docker-auth") {
		t.Errorf("expected docker-auth in lost section:\n%s", out)
	}
	if !strings.Contains(out, "new-failure") {
		t.Errorf("expected new-failure in gained section:\n%s", out)
	}
	if !strings.Contains(out, "## Gate Failures") {
		t.Errorf("expected Gate Failures section:\n%s", out)
	}
}

// --- WriteJSON ---

func TestWriteJSONProducesValidJSON(t *testing.T) {
	res := compare.Result{
		Pass:              true,
		GeneratedAt:       "2026-04-26T00:00:00Z",
		BaselineTotal:     5,
		CurrentTotal:      5,
		BaselineMatchRate: 0.60,
		CurrentMatchRate:  0.80,
		Delta:             0.20,
	}
	var buf bytes.Buffer
	if err := compare.WriteJSON(&buf, res); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}
	var decoded compare.Result
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("unmarshal WriteJSON output: %v", err)
	}
	if decoded.Pass != res.Pass {
		t.Errorf("Pass = %v, want %v", decoded.Pass, res.Pass)
	}
	if decoded.BaselineTotal != res.BaselineTotal {
		t.Errorf("BaselineTotal = %d, want %d", decoded.BaselineTotal, res.BaselineTotal)
	}
	if decoded.Delta != res.Delta {
		t.Errorf("Delta = %v, want %v", decoded.Delta, res.Delta)
	}
}

func TestWriteJSONIsIndented(t *testing.T) {
	res := compare.Result{Pass: true}
	var buf bytes.Buffer
	if err := compare.WriteJSON(&buf, res); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}
	// Indented JSON contains newlines
	if !strings.Contains(buf.String(), "\n") {
		t.Error("expected indented JSON output with newlines")
	}
}
