package config

import (
	_ "embed"
	"errors"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

//go:embed default_config.toml
var defaultConfigFile []byte

// Config holds all luma settings.
type Config struct {
	Steps     Steps     `toml:"steps"`
	Display   Display   `toml:"display"`
	Guardrails Guardrails `toml:"guardrails"`
	Theme     Theme     `toml:"theme"`
}

type Steps struct {
	Small  int `toml:"small"`
	Medium int `toml:"medium"`
	Large  int `toml:"large"`
}

type Display struct {
	RefreshIntervalMs int  `toml:"refresh_interval_ms"`
	ShowDisplayName   bool `toml:"show_display_name"`
}

type Guardrails struct {
	DebounceMs      int `toml:"debounce_ms"`
	MaxDdcutilProcs int `toml:"max_ddcutil_procs"`
	CommandTimeoutS int `toml:"command_timeout_s"`
}

type Theme struct {
	AccentColor string `toml:"accent_color"`
}

// Default returns safe default configuration.
func Default() Config {
	return Config{
		Steps: Steps{
			Small:  1,
			Medium: 5,
			Large:  10,
		},
		Display: Display{
			RefreshIntervalMs: 5000,
			ShowDisplayName:   true,
		},
		Guardrails: Guardrails{
			DebounceMs:      300,
			MaxDdcutilProcs: 1,
			CommandTimeoutS: 8,
		},
		Theme: Theme{
			AccentColor: "#9B59B6",
		},
	}
}

// clamp enforces guardrail value ranges.
func clamp(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// validate enforces guardrail clamping on a loaded config.
func validate(cfg *Config) {
	cfg.Guardrails.MaxDdcutilProcs = clamp(cfg.Guardrails.MaxDdcutilProcs, 1, 2)
	cfg.Guardrails.DebounceMs = clamp(cfg.Guardrails.DebounceMs, 100, 2000)
	cfg.Guardrails.CommandTimeoutS = clamp(cfg.Guardrails.CommandTimeoutS, 3, 30)
	if cfg.Display.RefreshIntervalMs != 0 {
		cfg.Display.RefreshIntervalMs = clamp(cfg.Display.RefreshIntervalMs, 2000, 1<<30)
	}
	if cfg.Steps.Small < 1 {
		cfg.Steps.Small = 1
	}
	if cfg.Steps.Medium < 1 {
		cfg.Steps.Medium = 1
	}
	if cfg.Steps.Large < 1 {
		cfg.Steps.Large = 1
	}
	if cfg.Theme.AccentColor == "" {
		cfg.Theme.AccentColor = "#A78BFA"
	}
}

// LoadOrDefault loads config from ~/.config/luma/config.toml.
// Missing file returns safe defaults. Bad values are clamped.
func LoadOrDefault() Config {
	cfg, _ := loadFromPath(defaultPath())
	return cfg
}

// Load loads config from the default path and returns any error.
func Load() (Config, error) {
	return loadFromPath(defaultPath())
}

// loadFromPath loads and validates config from a specific path.
func loadFromPath(path string) (Config, error) {
	cfg := Default()
	_, err := toml.DecodeFile(path, &cfg)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			writeDefaultConfig(path)
			return cfg, nil
		}
		return cfg, err
	}
	validate(&cfg)
	return cfg, nil
}

func defaultPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "luma", "config.toml")
}

func writeDefaultConfig(path string) {
	_ = os.MkdirAll(filepath.Dir(path), 0o755)
	_ = os.WriteFile(path, defaultConfigFile, 0o644)
}
