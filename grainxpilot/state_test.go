package grainxpilot

import "testing"

func TestNeedsHumanApproval(t *testing.T) {
	item := ManifestItem{State: ItemStateReadyForUpload, QA: QAResult{Score: 0.95}}

	cfg := DefaultConfig()
	if NeedsHumanApproval(cfg, []ManifestItem{item}) {
		t.Fatal("auto mode without flags should not require human approval")
	}

	cfg.Mode = ModeDryRun
	if !NeedsHumanApproval(cfg, []ManifestItem{item}) {
		t.Fatal("dry_run should require human approval")
	}

	cfg = DefaultConfig()
	cfg.RequireHumanApproval = true
	if !NeedsHumanApproval(cfg, []ManifestItem{item}) {
		t.Fatal("require_human_approval=true should require human approval")
	}

	cfg = DefaultConfig()
	item.QA.ForcedReview = true
	if !NeedsHumanApproval(cfg, []ManifestItem{item}) {
		t.Fatal("forced review item should require human approval")
	}
}

func TestNextRunStatus(t *testing.T) {
	items := []ManifestItem{{State: ItemStateReadyForUpload, QA: QAResult{Score: 0.99}}}

	cfg := DefaultConfig()
	status := NextRunStatus(cfg, items)
	if status != RunStatusReadyForUpload {
		t.Fatalf("expected %q, got %q", RunStatusReadyForUpload, status)
	}

	cfg.Mode = ModeDryRun
	status = NextRunStatus(cfg, items)
	if status != RunStatusAwaitingApproval {
		t.Fatalf("expected %q, got %q", RunStatusAwaitingApproval, status)
	}
}

func TestForceReviewReasons(t *testing.T) {
	item := ManifestItem{
		QA: QAResult{
			Score:                     0.49,
			PII:                       true,
			ForcedReview:              true,
			LowSegmentationConfidence: true,
		},
		DocPath: "",
	}
	reasons := ForceReviewReasons(item, 0.8)
	if len(reasons) < 4 {
		t.Fatalf("expected multiple force-review reasons, got %v", reasons)
	}
}
