package grainxpilot

import "testing"

func TestManifestValidate(t *testing.T) {
	m := Manifest{
		RunID:                 "run_20260417_001",
		Mode:                  ModeAuto,
		Status:                RunStatusReadyForUpload,
		BrowserAttachStrategy: AttachStrategyBrowserURL,
		Items: []ManifestItem{{
			RecordingID: "grain_rec_123",
			Slug:        "lesson-001",
			Title:       "Prompt Engineering Basics",
			DocPath:     "docs/lesson-001.md",
			CharCount:   38120,
			QA:          QAResult{Score: 0.93},
			State:       ItemStateReadyForUpload,
		}},
	}
	if err := m.Validate(); err != nil {
		t.Fatalf("manifest should validate: %v", err)
	}

	m.Items[0].DocPath = ""
	if err := m.Validate(); err == nil {
		t.Fatal("expected validation error for missing doc path")
	}
}
