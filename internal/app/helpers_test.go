package app

import (
	"strings"
	"testing"

	"faultline/internal/store"
)

// ── writeStoreInfoText ────────────────────────────────────────────────────────

func TestWriteStoreInfoTextDisabled(t *testing.T) {
	var b strings.Builder
	writeStoreInfoText(&b, store.Info{})
	if !strings.Contains(b.String(), "disabled") {
		t.Errorf("expected 'disabled' in output, got %q", b.String())
	}
}

func TestWriteStoreInfoTextBackendWithPath(t *testing.T) {
	var b strings.Builder
	writeStoreInfoText(&b, store.Info{Backend: "sqlite", Path: "/tmp/test.db"})
	got := b.String()
	if !strings.Contains(got, "sqlite") {
		t.Errorf("expected backend name 'sqlite', got %q", got)
	}
	if !strings.Contains(got, "/tmp/test.db") {
		t.Errorf("expected path '/tmp/test.db', got %q", got)
	}
}

func TestWriteStoreInfoTextBackendWithoutPath(t *testing.T) {
	var b strings.Builder
	writeStoreInfoText(&b, store.Info{Backend: "sqlite"})
	got := b.String()
	if !strings.Contains(got, "sqlite") {
		t.Errorf("expected backend name 'sqlite', got %q", got)
	}
}

func TestWriteStoreInfoTextEmptyBackendFallsBackToStoreLabel(t *testing.T) {
	// When Backend is empty but Path is set, the label should default to "store".
	var b strings.Builder
	writeStoreInfoText(&b, store.Info{Path: "/tmp/test.db"})
	got := b.String()
	if !strings.Contains(got, "store") {
		t.Errorf("expected 'store' label, got %q", got)
	}
	if !strings.Contains(got, "/tmp/test.db") {
		t.Errorf("expected path, got %q", got)
	}
}

func TestWriteStoreInfoTextDegradedShowsWarning(t *testing.T) {
	var b strings.Builder
	writeStoreInfoText(&b, store.Info{
		Backend:  "sqlite",
		Path:     "/tmp/test.db",
		Degraded: true,
		Warning:  "file is corrupt",
	})
	got := b.String()
	if !strings.Contains(got, "file is corrupt") {
		t.Errorf("expected warning message in output, got %q", got)
	}
}

// A degraded store without a warning message should not emit "Warning:".
func TestWriteStoreInfoTextDegradedWithoutWarningNoWarningLine(t *testing.T) {
	var b strings.Builder
	writeStoreInfoText(&b, store.Info{Backend: "sqlite", Degraded: true, Warning: ""})
	if strings.Contains(b.String(), "Warning:") {
		t.Errorf("unexpected Warning: line when warning is empty, got %q", b.String())
	}
}

// ── historyWindow ─────────────────────────────────────────────────────────────

func TestHistoryWindowInvalidFirstSeenAt(t *testing.T) {
	if got := historyWindow("not-a-date", "2026-04-22T10:00:00Z"); got != "" {
		t.Errorf("expected empty for invalid firstSeenAt, got %q", got)
	}
}

func TestHistoryWindowInvalidLastSeenAt(t *testing.T) {
	if got := historyWindow("2026-04-22T10:00:00Z", "not-a-date"); got != "" {
		t.Errorf("expected empty for invalid lastSeenAt, got %q", got)
	}
}

func TestHistoryWindowEndBeforeStart(t *testing.T) {
	if got := historyWindow("2026-04-22T12:00:00Z", "2026-04-22T10:00:00Z"); got != "" {
		t.Errorf("expected empty when end is before start, got %q", got)
	}
}

func TestHistoryWindowDays(t *testing.T) {
	got := historyWindow("2026-04-20T10:00:00Z", "2026-04-23T10:00:00Z")
	if got != "3d" {
		t.Errorf("expected '3d', got %q", got)
	}
}

func TestHistoryWindowHours(t *testing.T) {
	got := historyWindow("2026-04-22T10:00:00Z", "2026-04-22T13:00:00Z")
	if got != "3h" {
		t.Errorf("expected '3h', got %q", got)
	}
}

func TestHistoryWindowMinutes(t *testing.T) {
	got := historyWindow("2026-04-22T10:00:00Z", "2026-04-22T10:30:00Z")
	if got != "30m" {
		t.Errorf("expected '30m', got %q", got)
	}
}

func TestHistoryWindowSubMinuteReturnsEmpty(t *testing.T) {
	got := historyWindow("2026-04-22T10:00:00Z", "2026-04-22T10:00:30Z")
	if got != "" {
		t.Errorf("expected empty for sub-minute duration, got %q", got)
	}
}
