package tui

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

// UI helpers for consistent CLI presentation (no external deps, no flow changes).

func AppHeader(out io.Writer) {
	BlankLine(out)
	fmt.Fprintln(out, style("QUIBIT", ColorPrimary))
	fmt.Fprintln(out, style("Intelligent project generator for engineers.", ColorMuted))
	Divider(out)
}

func BlankLine(out io.Writer) {
	fmt.Fprintln(out, "")
}

func Divider(out io.Writer) {
	width := clampWidth(terminalWidth(out))
	line := strings.Repeat("─", width)
	fmt.Fprintln(out, style(line, ColorDivider))
}

func Heading(out io.Writer, title string) {
	BlankLine(out)
	title = strings.TrimSpace(title)
	if title == "" {
		return
	}
	fmt.Fprintln(out, style(title, ColorPrimary))
}

func Context(out io.Writer, text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	fmt.Fprintln(out, style(text, ColorBody))
}

func Hint(out io.Writer, text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	fmt.Fprintln(out, style(text, ColorMuted))
}

func DefaultValue(out io.Writer, text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	fmt.Fprintln(out, style("Default · "+text, ColorMuted))
}

func ControlsSelect(out io.Writer) {
	Hint(out, "↑/↓ navigate  ·  Enter select")
}

func Status(out io.Writer, text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	BlankLine(out)
	text = strings.TrimSuffix(text, "...")
	text = strings.TrimSuffix(text, ".")
	fmt.Fprintln(out, style("• "+text+"…", ColorStatus))
}

func Done(out io.Writer, text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	fmt.Fprintln(out, style("✓ "+text, ColorSuccess))
}

func PrintError(out io.Writer, context string, err error) {
	if err == nil {
		return
	}
	BlankLine(out)
	title := strings.TrimSpace(context)
	if title == "" {
		title = "Request failed"
	}
	fmt.Fprintln(out, style("ERROR · "+title, ColorErrorTitle))
	fmt.Fprintln(out, style(strings.TrimSpace(err.Error()), ColorErrorDetail))
}

func underlineWidth(s string, max int) int {
	if max <= 0 {
		return 0
	}
	n := utf8.RuneCountInString(s)
	if n <= 0 {
		return 0
	}
	if n > max {
		return max
	}
	return n
}
