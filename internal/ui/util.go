package ui

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

func presetRates() []string {
	return []string{"0.5×", "0.75×", "1.0×", "1.25×", "1.5×", "2.0×", "3.0×", "4.0×", "6.0×", "8.0×"}
}

func parseRate(s string) (float64, bool) {
	s = strings.TrimSuffix(strings.TrimSpace(s), "×")
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

func parseStep(s string) (int, bool) {
	s = strings.TrimSuffix(strings.TrimSpace(s), "s")
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0, false
	}
	return v, true
}

func clock(ms int64) string {
	if ms < 0 {
		ms = 0
	}
	total := ms / 1000
	h, m, s := total/3600, (total%3600)/60, total%60
	if h > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%02d:%02d", m, s)
}

func shortName(path string) string {
	base := filepath.Base(path)
	if len([]rune(base)) > 48 {
		r := []rune(base)
		return string(r[:45]) + "…"
	}
	return base
}
