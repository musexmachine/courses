package grainxpilot

import "testing"

func TestDefaultConfigIsValid(t *testing.T) {
	cfg := DefaultConfig()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("default config should validate: %v", err)
	}
	if cfg.Mode != ModeAuto {
		t.Fatalf("expected default mode %q, got %q", ModeAuto, cfg.Mode)
	}
	if cfg.BatchSize != 25 {
		t.Fatalf("expected default batch size 25, got %d", cfg.BatchSize)
	}
	if cfg.CharBudgetPerDoc != 42000 {
		t.Fatalf("expected default char budget 42000, got %d", cfg.CharBudgetPerDoc)
	}
}

func TestConfigValidateRejectsInvalidValues(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Mode = "bad"
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected invalid mode error")
	}

	cfg = DefaultConfig()
	cfg.BatchSize = 0
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected invalid batch size error")
	}

	cfg = DefaultConfig()
	cfg.CharBudgetPerDoc = 50001
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected invalid char budget error")
	}
}
