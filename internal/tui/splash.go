package tui

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"strings"
	"time"
	"unicode/utf8"
)

func ShowSplashScreen(ctx context.Context, in *os.File, out io.Writer, mode string) (shown bool, err error) {
	if in == nil || out == nil {
		return false, nil
	}
	if !isTerminal(int(in.Fd())) {
		return false, nil
	}
	type fdWriter interface{ Fd() uintptr }
	fw, ok := out.(fdWriter)
	if !ok || !isTerminal(int(fw.Fd())) {
		return false, nil
	}

	width := clampWidth(terminalWidth(out))

	title := "quibit"
	version := buildVersion()
	titleLine := title
	if version != "" {
		titleLine = title + "  " + style(version, ColorMuted)
	} else {
		titleLine = style(titleLine, ColorPrimary)
	}

	if version != "" {
		titleLine = style(title, ColorPrimary) + "  " + style(version, ColorMuted)
	}

	underline := strings.Repeat("─", underlineWidth(title, 18))
	tagline := style(splashTagline(mode, width), ColorMuted)
	createdBy := style("Created by", ColorMuted)
	author := style("Albert Mangiri", ColorGroupHeader)

	lines := []string{
		centerLine(titleLine, width),
		centerLine(style(underline, ColorDivider), width),
		centerLine(tagline, width),
		"",
		centerLine(createdBy, width),
		centerLine(author, width),
	}

	revealLines(ctx, out, lines)
	fmt.Fprintln(out, "")
	return true, nil
}

func splashTagline(mode string, width int) string {
	mode = strings.ToLower(strings.TrimSpace(mode))
	if mode == "" {
		mode = "idle"
	}

	var base string
	switch mode {
	case "generate":
		base = "Shaping a new project."
	case "continue":
		base = "Evolving an existing idea."
	case "browse":
		base = "Reviewing your work."
	default:
		base = "Engineering ideas, thoughtfully."
		mode = "idle"
	}

	if mode != "idle" && width >= 44 {
		base = base + style("  ·  "+mode, ColorDivider)
	}
	return base
}

func revealLines(ctx context.Context, out io.Writer, lines []string) {
	if len(lines) == 0 {
		return
	}
	if !motionAllowed(out) {
		for _, line := range lines {
			fmt.Fprintln(out, line)
		}
		return
	}

	const perLine = 70 * time.Millisecond
	for i := range lines {
		fmt.Fprintln(out, lines[i])

		if i == len(lines)-1 {
			break
		}
		timer := time.NewTimer(perLine)
		select {
		case <-ctx.Done():
			timer.Stop()

			for j := i + 1; j < len(lines); j++ {
				fmt.Fprintln(out, lines[j])
			}
			return
		case <-timer.C:
		}
	}
}

func centerLine(s string, width int) string {
	s = strings.TrimRight(s, "\r\n")
	if width <= 0 {
		return s
	}
	n := utf8.RuneCountInString(stripANSI(s))
	if n <= 0 || n >= width {
		return s
	}
	pad := (width - n) / 2
	if pad <= 0 {
		return s
	}
	return strings.Repeat(" ", pad) + s
}

func buildVersion() string {

	bi, ok := debug.ReadBuildInfo()
	if !ok || bi == nil {
		return ""
	}
	v := strings.TrimSpace(bi.Main.Version)
	if v == "" || v == "(devel)" {
		return ""
	}

	const max = 24
	if utf8.RuneCountInString(v) > max {
		rs := []rune(v)
		v = string(rs[:max-1]) + "…"
	}
	return v
}
