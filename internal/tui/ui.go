package tui

import (
	"fmt"
	"io"
	"strings"
)

func AppHeader(out io.Writer) {
	l := LayoutFor(out)
	writeSectionGap(out, l)
	
	// Simple colorful title without box
	titleText := style("Q", ColorNeonCyan) + 
	             style("U", ColorNeonBlue) +
	             style("I", ColorNeonPurple) +
	             style("B", ColorNeonMagenta) +
	             style("I", ColorDeepPurple) +
	             style("T", ColorNeonBlue)
	
	writeHeader(out, l, titleText)
	writeHeader(out, l, style("Intelligent project generator for engineers.", ColorMuted))
	
	// Divider with neon color
	HeaderDivider(out)
}

func BlankLine(out io.Writer) {
	fmt.Fprintln(out, "")
}

func Divider(out io.Writer) {
	l := LayoutFor(out)
	w := l.ContentWidth()
	line := strings.Repeat("━", w) // Heavy horizontal line
	fmt.Fprintln(out, leftPad(l.HPad())+style(line, ColorNeonCyan))
}

func HeaderDivider(out io.Writer) {
	l := LayoutFor(out)
	w := l.ContentWidth()
	line := style(strings.Repeat("─", w), ColorDivider)
	if l.CenterHeaders() {
		fmt.Fprintln(out, centerInBlock(line, l))
		return
	}
	fmt.Fprintln(out, leftPad(l.HPad())+line)
}

func Heading(out io.Writer, title string) {
	l := LayoutFor(out)
	writeSectionGap(out, l)
	title = strings.TrimSpace(title)
	if title == "" {
		return
	}
	// Add visual indicator
	writeLine(out, l, style("▸ ", ColorNeonCyan)+style(title, ColorNeonCyan))
}

func Context(out io.Writer, text string) {
	l := LayoutFor(out)
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	writeWrapped(out, l, text, ColorBody, false)
}

func Hint(out io.Writer, text string) {
	l := LayoutFor(out)
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	writeWrapped(out, l, text, ColorMuted, false)
}

func DefaultValue(out io.Writer, text string) {
	l := LayoutFor(out)
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	writeWrapped(out, l, "Default · "+text, ColorMuted, false)
}

func ControlsSelect(out io.Writer) {
	Hint(out, "↑/↓ navigate  ·  Enter select")
}

func Status(out io.Writer, text string) {
	l := LayoutFor(out)
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	writeSectionGap(out, l)
	text = strings.TrimSuffix(text, "...")
	text = strings.TrimSuffix(text, ".")
	writeWrapped(out, l, style("◆", ColorNeonBlue)+" "+text+"…", ColorStatus, false)
}

func Done(out io.Writer, text string) {
	l := LayoutFor(out)
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	writeWrapped(out, l, style("✓", ColorNeonGreen)+" "+text, ColorSuccess, false)
}

func PrintError(out io.Writer, context string, err error) {
	if err == nil {
		return
	}
	l := LayoutFor(out)
	writeSectionGap(out, l)
	title := strings.TrimSpace(context)
	if title == "" {
		title = "Request failed"
	}
	writeWrapped(out, l, "ERROR · "+title, ColorErrorTitle, false)
	writeWrapped(out, l, strings.TrimSpace(err.Error()), ColorErrorDetail, false)
}

func writeLine(out io.Writer, l Layout, line string) {
	fmt.Fprintln(out, leftPad(l.HPad())+line)
}

func writeHeader(out io.Writer, l Layout, line string) {
	if l.CenterHeaders() {
		fmt.Fprintln(out, centerInBlock(line, l))
		return
	}
	writeLine(out, l, line)
}

func writeWrapped(out io.Writer, l Layout, text string, sgr string, center bool) {
	w := l.ContentWidth()
	if w <= 0 {
		fmt.Fprintln(out, style(text, sgr))
		return
	}
	lines := wrapSoft(text, w)
	for _, ln := range lines {
		styled := style(ln, sgr)
		if center {
			fmt.Fprintln(out, centerInBlock(styled, l))
			continue
		}
		fmt.Fprintln(out, leftPad(l.HPad())+styled)
	}
}

func writeSectionGap(out io.Writer, l Layout) {
	for i := 0; i < l.VSectionLines(); i++ {
		fmt.Fprintln(out, "")
	}
}

func centerInBlock(s string, l Layout) string {
	w := l.ContentWidth()
	if w <= 0 {
		return s
	}
	n := visibleRuneLen(s)
	if n <= 0 || n >= w {
		return leftPad(l.HPad()) + s
	}
	pad := (w - n) / 2
	return leftPad(l.HPad()+pad) + s
}
