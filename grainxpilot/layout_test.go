package grainxpilot

import (
	"path/filepath"
	"testing"
	"time"
)

func TestNewRunLayout(t *testing.T) {
	ts := time.Date(2026, 4, 17, 21, 45, 0, 0, time.UTC)
	layout, err := NewRunLayout("/tmp/runs", ts, "run_20260417_001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wantRunDir := filepath.Join("/tmp/runs", "2026-04-17", "run_20260417_001")
	if layout.RunDir != wantRunDir {
		t.Fatalf("expected run dir %q, got %q", wantRunDir, layout.RunDir)
	}
	if got := layout.DocPathForSlug("lesson-001"); got != filepath.Join(wantRunDir, "docs", "lesson-001.md") {
		t.Fatalf("unexpected doc path: %q", got)
	}
}
