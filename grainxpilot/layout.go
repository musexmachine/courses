package grainxpilot

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type RunLayout struct {
	Root         string
	DateDir      string
	RunID        string
	RunDir       string
	ManifestPath string
	QueuePath    string
	ReviewPath   string
	DocsDir      string
	SourceDir    string
	ExportsDir   string
	LogsDir      string
}

func NewRunLayout(root string, ts time.Time, runID string) (RunLayout, error) {
	if root == "" {
		return RunLayout{}, fmt.Errorf("root is required")
	}
	if runID == "" {
		return RunLayout{}, fmt.Errorf("runID is required")
	}
	dateDir := ts.Format("2006-01-02")
	runDir := filepath.Join(root, dateDir, runID)
	return RunLayout{
		Root:         root,
		DateDir:      dateDir,
		RunID:        runID,
		RunDir:       runDir,
		ManifestPath: filepath.Join(runDir, "manifest.json"),
		QueuePath:    filepath.Join(runDir, "queue.json"),
		ReviewPath:   filepath.Join(runDir, "review.md"),
		DocsDir:      filepath.Join(runDir, "docs"),
		SourceDir:    filepath.Join(runDir, "source"),
		ExportsDir:   filepath.Join(runDir, "exports"),
		LogsDir:      filepath.Join(runDir, "logs"),
	}, nil
}

func (l RunLayout) EnsureDirs() error {
	for _, dir := range []string{l.RunDir, l.DocsDir, l.SourceDir, l.ExportsDir, l.LogsDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return nil
}

func (l RunLayout) DocPathForSlug(slug string) string {
	return filepath.Join(l.DocsDir, slug+".md")
}
