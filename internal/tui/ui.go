package tui

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

// UI helpers for consistent CLI presentation (no external deps, no flow changes).

func BlankLine(out io.Writer) {
	fmt.Fprintln(out, "")
}

func Divider(out io.Writer) {
	width := terminalWidth(out)
	if width <= 0 {
		width = 80
	}
	// Keep dividers readable on wide terminals.
	if width > 84 {
		width = 84
	}
	if width < 24 {
		width = 24
	}
	fmt.Fprintln(out, strings.Repeat("-", width))
}

func Heading(out io.Writer, title string) {
	BlankLine(out)
	fmt.Fprintln(out, title)
	u := underlineWidth(title, 42)
	if u > 0 {
		fmt.Fprintln(out, strings.Repeat("-", u))
	}
}

func Context(out io.Writer, text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	fmt.Fprintln(out, text)
}

func Hint(out io.Writer, text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	fmt.Fprintf(out, "Hint: %s\n", text)
}

func DefaultValue(out io.Writer, text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	fmt.Fprintf(out, "Default: %s\n", text)
}

func ControlsSelect(out io.Writer) {
	fmt.Fprintln(out, "Controls: ↑/↓ to navigate • Enter to confirm")
}

func Status(out io.Writer, text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	BlankLine(out)
	fmt.Fprintf(out, "Status: %s...\n", text)
}

func Done(out io.Writer, text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	fmt.Fprintf(out, "Done: %s\n", text)
}

func PrintError(out io.Writer, context string, err error) {
	if err == nil {
		return
	}
	context = strings.TrimSpace(context)
	if context == "" {
		fmt.Fprintf(out, "Error: %s\n", strings.TrimSpace(err.Error()))
		return
	}
	fmt.Fprintf(out, "Error: %s. %s\n", context, strings.TrimSpace(err.Error()))
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

