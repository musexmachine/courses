package grainxpilot

import (
	"fmt"
	"strings"
)

type RunStatus string

const (
	RunStatusDiscovered           RunStatus = "DISCOVERED"
	RunStatusTranscriptsFetched   RunStatus = "TRANSCRIPTS_FETCHED"
	RunStatusNormalized           RunStatus = "NORMALIZED"
	RunStatusBatchFolderReady     RunStatus = "BATCH_FOLDER_READY"
	RunStatusAwaitingApproval     RunStatus = "AWAITING_APPROVAL"
	RunStatusReadyForUpload       RunStatus = "READY_FOR_UPLOAD"
	RunStatusBrowserAttached      RunStatus = "BROWSER_ATTACHED"
	RunStatusAuthVerified         RunStatus = "AUTH_VERIFIED"
	RunStatusXPilotUploadProgress RunStatus = "XPILOT_UPLOAD_IN_PROGRESS"
	RunStatusXPilotParseComplete  RunStatus = "XPILOT_PARSE_COMPLETE"
	RunStatusXPilotRenderProgress RunStatus = "XPILOT_RENDER_IN_PROGRESS"
	RunStatusXPilotExportReady    RunStatus = "XPILOT_EXPORT_READY"
	RunStatusAssetsDownloaded     RunStatus = "ASSETS_DOWNLOADED"
	RunStatusAuthRequired         RunStatus = "AUTH_REQUIRED"
	RunStatusComplete             RunStatus = "COMPLETE"
)

type ItemState string

const (
	ItemStateNew            ItemState = "NEW"
	ItemStateFetched        ItemState = "FETCHED"
	ItemStateNormalized     ItemState = "NORMALIZED"
	ItemStateDocReady       ItemState = "DOC_READY"
	ItemStateReviewRequired ItemState = "REVIEW_REQUIRED"
	ItemStateReadyForUpload ItemState = "READY_FOR_UPLOAD"
	ItemStateUploaded       ItemState = "UPLOADED"
	ItemStateParsed         ItemState = "PARSED"
	ItemStateRendered       ItemState = "RENDERED"
	ItemStateExported       ItemState = "EXPORTED"
	ItemStateDownloaded     ItemState = "DOWNLOADED"
	ItemStateFailed         ItemState = "FAILED"
)

type QAResult struct {
	Score                     float64 `json:"score"`
	PII                       bool    `json:"pii"`
	ForcedReview              bool    `json:"forced_review"`
	LowSegmentationConfidence bool    `json:"low_segmentation_confidence"`
}

type ArtifactPaths struct {
	MP4   string `json:"mp4,omitempty"`
	SCORM string `json:"scorm,omitempty"`
}

type ManifestItem struct {
	RecordingID string        `json:"recording_id"`
	Slug        string        `json:"slug"`
	Title       string        `json:"title"`
	DocPath     string        `json:"doc_path"`
	CharCount   int           `json:"char_count"`
	QA          QAResult      `json:"qa"`
	State       ItemState     `json:"state"`
	Artifacts   ArtifactPaths `json:"artifacts"`
}

type Manifest struct {
	RunID                 string                `json:"run_id"`
	Mode                  Mode                  `json:"mode"`
	Status                RunStatus             `json:"status"`
	RequireHumanApproval  bool                  `json:"require_human_approval"`
	BrowserAttachStrategy BrowserAttachStrategy `json:"browser_attach_strategy"`
	Items                 []ManifestItem        `json:"items"`
}

func (m Manifest) Validate() error {
	if strings.TrimSpace(m.RunID) == "" {
		return fmt.Errorf("run id is required")
	}
	switch m.Mode {
	case ModeAuto, ModeDryRun:
	default:
		return fmt.Errorf("invalid manifest mode: %q", m.Mode)
	}
	if err := validateAttachStrategy(m.BrowserAttachStrategy); err != nil {
		return err
	}
	if len(m.Items) == 0 {
		return fmt.Errorf("manifest must contain at least one item")
	}
	for i, item := range m.Items {
		if strings.TrimSpace(item.RecordingID) == "" {
			return fmt.Errorf("item %d missing recording id", i)
		}
		if strings.TrimSpace(item.Slug) == "" {
			return fmt.Errorf("item %d missing slug", i)
		}
		if strings.TrimSpace(item.Title) == "" {
			return fmt.Errorf("item %d missing title", i)
		}
		if strings.TrimSpace(item.DocPath) == "" {
			return fmt.Errorf("item %d missing doc path", i)
		}
		if item.CharCount <= 0 || item.CharCount > maxXPilotCharsPerVideo {
			return fmt.Errorf("item %d invalid char count %d", i, item.CharCount)
		}
		if item.QA.Score < 0 || item.QA.Score > 1 {
			return fmt.Errorf("item %d invalid QA score %.2f", i, item.QA.Score)
		}
		switch item.State {
		case ItemStateNew, ItemStateFetched, ItemStateNormalized, ItemStateDocReady, ItemStateReviewRequired,
			ItemStateReadyForUpload, ItemStateUploaded, ItemStateParsed, ItemStateRendered, ItemStateExported,
			ItemStateDownloaded, ItemStateFailed:
		default:
			return fmt.Errorf("item %d invalid state %q", i, item.State)
		}
	}
	return nil
}

type Recording struct {
	ID    string            `json:"id"`
	Title string            `json:"title"`
	Meta  map[string]string `json:"meta,omitempty"`
}

type TranscriptSegment struct {
	ParticipantID string `json:"participant_id,omitempty"`
	Speaker       string `json:"speaker,omitempty"`
	StartMS       int64  `json:"start_ms"`
	EndMS         int64  `json:"end_ms"`
	Text          string `json:"text"`
}

type Hook struct {
	ID       string                 `json:"id"`
	HookType string                 `json:"hook_type"`
	URL      string                 `json:"url"`
	Meta     map[string]interface{} `json:"meta,omitempty"`
}
