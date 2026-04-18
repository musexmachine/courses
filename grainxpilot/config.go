package grainxpilot

import "fmt"

const maxXPilotCharsPerVideo = 50000

type Mode string

const (
	ModeAuto   Mode = "auto"
	ModeDryRun Mode = "dry_run"
)

type BrowserAttachStrategy string

const (
	AttachStrategyBrowserURL        BrowserAttachStrategy = "browserUrl"
	AttachStrategyAutoConnect       BrowserAttachStrategy = "autoConnect"
	AttachStrategyPlaywrightCDP     BrowserAttachStrategy = "playwright_cdp"
	AttachStrategyPuppeteerConnect  BrowserAttachStrategy = "puppeteer_connect"
	AttachStrategyChromeDebuggerExt BrowserAttachStrategy = "chrome_debugger_extension"
)

type ExportFormat string

const (
	ExportFormatMP4   ExportFormat = "mp4"
	ExportFormatSCORM ExportFormat = "scorm"
)

type Config struct {
	Mode                  Mode
	RequireHumanApproval  bool
	PauseBeforeUpload     bool
	PauseBeforeRender     bool
	PrimaryAttachStrategy BrowserAttachStrategy
	FallbackStrategies    []BrowserAttachStrategy
	BatchSize             int
	CharBudgetPerDoc      int
	ExportFormats         []ExportFormat
	MinQAScore            float64
}

func DefaultConfig() Config {
	return Config{
		Mode:                  ModeAuto,
		RequireHumanApproval:  false,
		PauseBeforeUpload:     false,
		PauseBeforeRender:     false,
		PrimaryAttachStrategy: AttachStrategyBrowserURL,
		FallbackStrategies: []BrowserAttachStrategy{
			AttachStrategyAutoConnect,
			AttachStrategyPlaywrightCDP,
			AttachStrategyPuppeteerConnect,
			AttachStrategyChromeDebuggerExt,
		},
		BatchSize:        25,
		CharBudgetPerDoc: 42000,
		ExportFormats:    []ExportFormat{ExportFormatMP4, ExportFormatSCORM},
		MinQAScore:       0.80,
	}
}

func (c Config) Validate() error {
	switch c.Mode {
	case ModeAuto, ModeDryRun:
	default:
		return fmt.Errorf("invalid mode: %q", c.Mode)
	}
	if c.BatchSize <= 0 {
		return fmt.Errorf("batch size must be > 0")
	}
	if c.CharBudgetPerDoc <= 0 || c.CharBudgetPerDoc > maxXPilotCharsPerVideo {
		return fmt.Errorf("char budget per doc must be between 1 and %d", maxXPilotCharsPerVideo)
	}
	if err := validateAttachStrategy(c.PrimaryAttachStrategy); err != nil {
		return err
	}
	for _, s := range c.FallbackStrategies {
		if err := validateAttachStrategy(s); err != nil {
			return err
		}
	}
	if c.MinQAScore < 0 || c.MinQAScore > 1 {
		return fmt.Errorf("min QA score must be between 0 and 1")
	}
	for _, f := range c.ExportFormats {
		switch f {
		case ExportFormatMP4, ExportFormatSCORM:
		default:
			return fmt.Errorf("invalid export format: %q", f)
		}
	}
	return nil
}

func validateAttachStrategy(s BrowserAttachStrategy) error {
	switch s {
	case AttachStrategyBrowserURL,
		AttachStrategyAutoConnect,
		AttachStrategyPlaywrightCDP,
		AttachStrategyPuppeteerConnect,
		AttachStrategyChromeDebuggerExt:
		return nil
	default:
		return fmt.Errorf("invalid browser attach strategy: %q", s)
	}
}
