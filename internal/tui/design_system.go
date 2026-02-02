package tui

import (
	"os"
	"strings"
)

// Spacing system (vertical rhythm).
const (
	SpaceXS = 0 // no blank lines
	SpaceSM = 1 // within a group
	SpaceMD = 2 // between sections
	SpaceLG = 3 // between screens/states
)

// Semantic color tokens (SGR fragments; applied via style()).
const (
	ColorPrimary     = "1;38;5;255"
	ColorBody        = "38;5;252"
	ColorMuted       = "38;5;245"
	ColorDivider     = "38;5;238"
	ColorAccent      = "1;38;5;81"
	ColorStatus      = "38;5;111"
	ColorSuccess     = "38;5;114"
	ColorWarning     = "38;5;214"
	ColorErrorTitle  = "1;38;5;203"
	ColorErrorDetail = "38;5;245"
)

const ansiReset = "\033[0m"

func ansi(sgr string) string { return "\033[" + sgr + "m" }

func noColor() bool {
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return true
	}
	return false
}

func style(s, sgr string) string {
	if noColor() || strings.TrimSpace(sgr) == "" {
		return s
	}
	return ansi(sgr) + s + ansiReset
}

func clampWidth(w int) int {
	if w <= 0 {
		return 80
	}
	if w > 96 {
		return 96
	}
	if w < 28 {
		return 28
	}
	return w
}
