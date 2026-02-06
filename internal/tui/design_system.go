package tui

import (
	"os"
	"strings"
)

const (
	SpaceXS = 0
	SpaceSM = 1
	SpaceMD = 2
	SpaceLG = 3
)

const (
	// Base colors - clean and crisp
	ColorPrimary = "1;38;5;255" // Bright white
	ColorBody    = "38;5;252"   // Light gray
	ColorMuted   = "38;5;245"   // Medium gray
	ColorDivider = "38;5;238"   // Dark gray

	// Blue-Purple theme colors - vibrant and cohesive
	ColorNeonCyan    = "1;38;5;51"  // Bright cyan (primary accent)
	ColorNeonBlue    = "1;38;5;39"  // Electric blue
	ColorNeonPurple  = "1;38;5;141" // Bright purple
	ColorNeonMagenta = "1;38;5;201" // Bright magenta (highlight)
	ColorDeepBlue    = "1;38;5;33"  // Deep blue
	ColorDeepPurple  = "1;38;5;93"  // Deep purple
	
	// Legacy colors for compatibility
	ColorNeonGreen   = "1;38;5;46"  // Bright green (success)
	ColorNeonYellow  = "1;38;5;226" // Bright yellow (warning)

	// Semantic colors - blue-purple theme
	ColorAccent      = "1;38;5;51"  // Neon cyan
	ColorGroupHeader = "1;38;5;141" // Neon purple
	ColorStatus      = "38;5;111"   // Softer blue
	ColorSuccess     = "1;38;5;51"  // Cyan (instead of green)
	ColorWarning     = "1;38;5;226" // Yellow
	ColorErrorTitle  = "1;38;5;203" // Bright red
	ColorErrorDetail = "38;5;245"   // Muted gray

	// Border colors - blue-purple theme
	ColorBorderPrimary   = "1;38;5;51"  // Neon cyan
	ColorBorderSecondary = "38;5;81"    // Medium cyan
	ColorBorderMuted     = "38;5;245"   // Gray
	ColorBorderAccent    = "1;38;5;201" // Neon magenta
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
