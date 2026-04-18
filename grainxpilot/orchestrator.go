package grainxpilot

import "time"

func BuildManifest(cfg Config, runID string, items []ManifestItem) Manifest {
	return Manifest{
		RunID:                 runID,
		Mode:                  cfg.Mode,
		Status:                NextRunStatus(cfg, items),
		RequireHumanApproval:  NeedsHumanApproval(cfg, items),
		BrowserAttachStrategy: cfg.PrimaryAttachStrategy,
		Items:                 items,
	}
}

func NewDefaultRunLayout(root, runID string) (RunLayout, error) {
	return NewRunLayout(root, time.Now().UTC(), runID)
}
