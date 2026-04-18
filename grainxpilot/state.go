package grainxpilot

import "fmt"

func NeedsHumanApproval(cfg Config, items []ManifestItem) bool {
	if cfg.Mode == ModeDryRun || cfg.RequireHumanApproval || cfg.PauseBeforeUpload || cfg.PauseBeforeRender {
		return true
	}
	for _, item := range items {
		if item.QA.ForcedReview {
			return true
		}
	}
	return false
}

func NextRunStatus(cfg Config, items []ManifestItem) RunStatus {
	if NeedsHumanApproval(cfg, items) {
		return RunStatusAwaitingApproval
	}
	return RunStatusReadyForUpload
}

func ForceReviewReasons(item ManifestItem, minQAScore float64) []string {
	var reasons []string
	if item.QA.Score < minQAScore {
		reasons = append(reasons, fmt.Sprintf("qa score %.2f below threshold %.2f", item.QA.Score, minQAScore))
	}
	if item.QA.PII {
		reasons = append(reasons, "PII/redaction issue detected")
	}
	if item.QA.ForcedReview {
		reasons = append(reasons, "forced review flag set")
	}
	if item.QA.LowSegmentationConfidence {
		reasons = append(reasons, "low segmentation confidence")
	}
	if item.DocPath == "" {
		reasons = append(reasons, "missing doc path")
	}
	return reasons
}
