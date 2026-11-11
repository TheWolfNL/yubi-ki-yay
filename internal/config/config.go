// Package config loads and saves the Yubi Ki-Yay settings file.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const systemSoundsDir = "/System/Library/Sounds"

// Config is shared by the menu-bar app and the notify CLI.
type Config struct {
	Banner       bool   `json:"banner"`
	Sound        bool   `json:"sound"`
	SoundName    string `json:"soundName"`
	TerminalLine bool   `json:"terminalLine"`
	TrayFlash    bool   `json:"trayFlash"`
	TrayMode     string `json:"trayMode"`
}

// Tray icon visibility modes.
const (
	TrayAlways  = "always"
	TrayOnTouch = "onTouch"
)

// Default returns the settings used when no config file exists.
func Default() Config {
	return Config{
		Banner:       true,
		Sound:        true,
		SoundName:    "Glass",
		TerminalLine: true,
		TrayFlash:    true,
		TrayMode:     TrayAlways,
	}
}

// TrayHidden reports whether the tray icon should stay hidden while idle.
func (c Config) TrayHidden() bool {
	return c.TrayMode == TrayOnTouch
}

// Dir returns ~/.config/yubi-ki-yay.
func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}
	return filepath.Join(home, ".config", "yubi-ki-yay"), nil
}

func path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// Load reads the config file, falling back to Default when it does not exist.
func Load() (Config, error) {
	p, err := path()
	if err != nil {
		return Default(), err
	}
	data, err := os.ReadFile(p)
	if errors.Is(err, fs.ErrNotExist) {
		return Default(), nil
	}
	if err != nil {
		return Default(), fmt.Errorf("read config: %w", err)
	}
	cfg := Default()
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Default(), fmt.Errorf("parse %s: %w", p, err)
	}
	return cfg, nil
}

// Save writes via temp file and rename because the CLI may read it concurrently.
func Save(cfg Config) error {
	p, err := path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	if err := os.Rename(tmp, p); err != nil {
		return fmt.Errorf("replace config: %w", err)
	}
	return nil
}

// SoundPath maps a system sound name such as "Glass" to its file.
func SoundPath(name string) string {
	return filepath.Join(systemSoundsDir, name+".aiff")
}

// SystemSounds lists the sound names available in /System/Library/Sounds.
func SystemSounds() ([]string, error) {
	entries, err := os.ReadDir(systemSoundsDir)
	if err != nil {
		return nil, fmt.Errorf("list system sounds: %w", err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if name, ok := strings.CutSuffix(e.Name(), ".aiff"); ok {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names, nil
}
