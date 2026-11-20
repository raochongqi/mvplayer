package ui

import (
	"encoding/json"
	"os"
)

// Settings is the small user-tweakable surface: the optional custom speed and
// seek-step lists, and whether history is recorded.
type Settings struct {
	Rates   []string `json:"rates"`
	Steps   []string `json:"steps"`
	History bool     `json:"history"`
}

var defaultSettings = Settings{
	Rates:   []string{"0.5×", "0.75×", "1.0×", "1.25×", "1.5×", "2.0×", "3.0×", "4.0×", "6.0×", "8.0×"},
	Steps:   []string{"10s", "30s", "60s", "120s"},
	History: true,
}

// LoadSettings reads settings, falling back to defaults on any error.
func LoadSettings(path string) Settings {
	s := defaultSettings
	b, err := os.ReadFile(path)
	if err != nil {
		return s
	}
	if err := json.Unmarshal(b, &s); err != nil {
		return defaultSettings
	}
	if len(s.Rates) == 0 {
		s.Rates = defaultSettings.Rates
	}
	if len(s.Steps) == 0 {
		s.Steps = defaultSettings.Steps
	}
	return s
}

// Save writes settings atomically.
func (s Settings) Save(path string) error {
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func (s Settings) rateChoices() []string { return append([]string(nil), s.Rates...) }
func (s Settings) stepChoices() []string { return append([]string(nil), s.Steps...) }
