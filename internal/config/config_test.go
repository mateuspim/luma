package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()
	if cfg.Guardrails.MaxDdcutilProcs != 1 {
		t.Errorf("MaxDdcutilProcs default = %d, want 1", cfg.Guardrails.MaxDdcutilProcs)
	}
	if cfg.Guardrails.DebounceMs != 300 {
		t.Errorf("DebounceMs default = %d, want 300", cfg.Guardrails.DebounceMs)
	}
	if cfg.Guardrails.CommandTimeoutS != 8 {
		t.Errorf("CommandTimeoutS default = %d, want 8", cfg.Guardrails.CommandTimeoutS)
	}
	if cfg.Display.RefreshIntervalMs != 5000 {
		t.Errorf("RefreshIntervalMs default = %d, want 5000", cfg.Display.RefreshIntervalMs)
	}
}

func TestValidate_Clamping(t *testing.T) {
	cfg := Default()
	cfg.Guardrails.MaxDdcutilProcs = 10  // above max
	cfg.Guardrails.DebounceMs = 10       // below min
	cfg.Guardrails.CommandTimeoutS = 100 // above max
	cfg.Display.RefreshIntervalMs = 100  // below min (non-zero)

	validate(&cfg)

	if cfg.Guardrails.MaxDdcutilProcs != 2 {
		t.Errorf("MaxDdcutilProcs = %d, want 2", cfg.Guardrails.MaxDdcutilProcs)
	}
	if cfg.Guardrails.DebounceMs != 100 {
		t.Errorf("DebounceMs = %d, want 100", cfg.Guardrails.DebounceMs)
	}
	if cfg.Guardrails.CommandTimeoutS != 30 {
		t.Errorf("CommandTimeoutS = %d, want 30", cfg.Guardrails.CommandTimeoutS)
	}
	if cfg.Display.RefreshIntervalMs != 2000 {
		t.Errorf("RefreshIntervalMs = %d, want 2000", cfg.Display.RefreshIntervalMs)
	}
}

func TestValidate_ZeroRefreshAllowed(t *testing.T) {
	cfg := Default()
	cfg.Display.RefreshIntervalMs = 0
	validate(&cfg)
	if cfg.Display.RefreshIntervalMs != 0 {
		t.Errorf("RefreshIntervalMs should remain 0 (disabled), got %d", cfg.Display.RefreshIntervalMs)
	}
}

func TestLoadFromPath_MissingFile(t *testing.T) {
	cfg, err := loadFromPath("/nonexistent/path/config.toml")
	if err != nil {
		t.Errorf("missing file should not return error, got: %v", err)
	}
	// Should return defaults
	def := Default()
	if cfg.Guardrails.DebounceMs != def.Guardrails.DebounceMs {
		t.Errorf("expected default DebounceMs %d, got %d", def.Guardrails.DebounceMs, cfg.Guardrails.DebounceMs)
	}
}

func TestLoadFromPath_PartialOverride(t *testing.T) {
	// Write a partial TOML file
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	content := `
[guardrails]
debounce_ms = 500
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := loadFromPath(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Overridden value
	if cfg.Guardrails.DebounceMs != 500 {
		t.Errorf("DebounceMs = %d, want 500", cfg.Guardrails.DebounceMs)
	}
	// Non-overridden values should be defaults
	if cfg.Steps.Small != 1 {
		t.Errorf("Steps.Small = %d, want 1 (default)", cfg.Steps.Small)
	}
	if cfg.Theme.AccentColor != "#9B59B6" {
		t.Errorf("AccentColor = %q, want default", cfg.Theme.AccentColor)
	}
}

func TestLoadFromPath_BadValues(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	content := `
[guardrails]
max_ddcutil_procs = 99
debounce_ms = 1
command_timeout_s = 999
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := loadFromPath(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// All values clamped
	if cfg.Guardrails.MaxDdcutilProcs != 2 {
		t.Errorf("MaxDdcutilProcs = %d, want 2", cfg.Guardrails.MaxDdcutilProcs)
	}
	if cfg.Guardrails.DebounceMs != 100 {
		t.Errorf("DebounceMs = %d, want 100", cfg.Guardrails.DebounceMs)
	}
	if cfg.Guardrails.CommandTimeoutS != 30 {
		t.Errorf("CommandTimeoutS = %d, want 30", cfg.Guardrails.CommandTimeoutS)
	}
}
