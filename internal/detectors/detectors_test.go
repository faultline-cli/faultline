package detectors

import (
	"strings"
	"testing"

	"faultline/internal/model"
)

// stubDetector is a minimal Detector used to exercise the registry.
type stubDetector struct {
	kind Kind
}

func (s stubDetector) Kind() Kind { return s.kind }
func (s stubDetector) Detect(_ []model.Playbook, _ Target) []model.Result {
	return nil
}

// ── Registry ─────────────────────────────────────────────────────────────────

func TestNewRegistryAndLookup(t *testing.T) {
	logDet := stubDetector{kind: KindLog}
	srcDet := stubDetector{kind: KindSource}
	reg := NewRegistry(logDet, srcDet)

	got, ok := reg.Lookup(KindLog)
	if !ok {
		t.Fatal("expected KindLog to be registered")
	}
	if got.Kind() != KindLog {
		t.Fatalf("expected KindLog, got %q", got.Kind())
	}

	got, ok = reg.Lookup(KindSource)
	if !ok {
		t.Fatal("expected KindSource to be registered")
	}
	if got.Kind() != KindSource {
		t.Fatalf("expected KindSource, got %q", got.Kind())
	}
}

func TestRegistryLookupMissing(t *testing.T) {
	reg := NewRegistry()
	_, ok := reg.Lookup(KindLog)
	if ok {
		t.Error("expected missing lookup to return false")
	}
}

func TestRegistryMustLookupError(t *testing.T) {
	reg := NewRegistry()
	_, err := reg.MustLookup(KindLog)
	if err == nil {
		t.Error("expected error for missing KindLog")
	}
}

func TestRegistryMustLookupSuccess(t *testing.T) {
	reg := NewRegistry(stubDetector{kind: KindLog})
	det, err := reg.MustLookup(KindLog)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if det.Kind() != KindLog {
		t.Fatalf("unexpected kind %q", det.Kind())
	}
}

// ── FilterPlaybooks ───────────────────────────────────────────────────────────

func TestFilterPlaybooksLog(t *testing.T) {
	pbs := []model.Playbook{
		{ID: "log-a", Detector: "log"},
		{ID: "src-a", Detector: "source"},
		{ID: "log-b"}, // empty detector defaults to log
		{ID: "src-b", Detector: "source"},
	}

	got := FilterPlaybooks(pbs, KindLog)
	if len(got) != 2 {
		t.Fatalf("expected 2 log playbooks, got %d", len(got))
	}
	for _, pb := range got {
		if pb.Detector != "" && pb.Detector != "log" {
			t.Errorf("unexpected detector %q in log filter result", pb.Detector)
		}
	}
}

func TestFilterPlaybooksSource(t *testing.T) {
	pbs := []model.Playbook{
		{ID: "log-a", Detector: "log"},
		{ID: "src-a", Detector: "source"},
		{ID: "src-b", Detector: "source"},
	}

	got := FilterPlaybooks(pbs, KindSource)
	if len(got) != 2 {
		t.Fatalf("expected 2 source playbooks, got %d", len(got))
	}
}

func TestFilterPlaybooksEmpty(t *testing.T) {
	got := FilterPlaybooks(nil, KindLog)
	if len(got) != 0 {
		t.Fatalf("expected empty result for nil input, got %d", len(got))
	}
}

func TestFilterPlaybooksNoneMatch(t *testing.T) {
	pbs := []model.Playbook{
		{ID: "src-a", Detector: "source"},
	}
	got := FilterPlaybooks(pbs, KindLog)
	if len(got) != 0 {
		t.Fatalf("expected 0 results, got %d", len(got))
	}
}

func TestRegistryMustLookupErrorIncludesAvailableKinds(t *testing.T) {
	reg := NewRegistry(stubDetector{kind: KindSource})
	_, err := reg.MustLookup(KindLog)
	if err == nil {
		t.Fatal("expected error for missing KindLog")
	}
	msg := err.Error()
	if !strings.Contains(msg, string(KindSource)) {
		t.Errorf("expected error message to include available kind %q, got: %s", KindSource, msg)
	}
}

func TestRegistryAvailableKindsEmpty(t *testing.T) {
	reg := NewRegistry()
	_, err := reg.MustLookup(KindLog)
	if err == nil {
		t.Fatal("expected error for empty registry")
	}
	if !strings.Contains(err.Error(), "none") {
		t.Errorf("expected 'none' in error for empty registry, got: %s", err.Error())
	}
}
