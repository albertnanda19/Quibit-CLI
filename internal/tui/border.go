package tui

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

// Border characters using Unicode box-drawing
type BorderStyle struct {
	TopLeft     string
	TopRight    string
	BottomLeft  string
	BottomRight string
	Horizontal  string
	Vertical    string
	TeeLeft     string
	TeeRight    string
	TeeTop      string
	TeeBottom   string
}

var (
	// BorderSingle uses single-line box-drawing characters
	BorderSingle = BorderStyle{
		TopLeft:     "┌",
		TopRight:    "┐",
		BottomLeft:  "└",
		BottomRight: "┘",
		Horizontal:  "─",
		Vertical:    "│",
		TeeLeft:     "├",
		TeeRight:    "┤",
		TeeTop:      "┬",
		TeeBottom:   "┴",
	}

	// BorderDouble uses double-line box-drawing characters
	BorderDouble = BorderStyle{
		TopLeft:     "╔",
		TopRight:    "╗",
		BottomLeft:  "╚",
		BottomRight: "╝",
		Horizontal:  "═",
		Vertical:    "║",
		TeeLeft:     "╠",
		TeeRight:    "╣",
		TeeTop:      "╦",
		TeeBottom:   "╩",
	}

	// BorderHeavy uses heavy-line box-drawing characters
	BorderHeavy = BorderStyle{
		TopLeft:     "┏",
		TopRight:    "┓",
		BottomLeft:  "┗",
		BottomRight: "┛",
		Horizontal:  "━",
		Vertical:    "┃",
		TeeLeft:     "┣",
		TeeRight:    "┫",
		TeeTop:      "┳",
		TeeBottom:   "┻",
	}

	// BorderRounded uses rounded corners
	BorderRounded = BorderStyle{
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "╰",
		BottomRight: "╯",
		Horizontal:  "─",
		Vertical:    "│",
		TeeLeft:     "├",
		TeeRight:    "┤",
		TeeTop:      "┬",
		TeeBottom:   "┴",
	}
)

// Box represents a bordered content area
type Box struct {
	Style       BorderStyle
	Width       int
	ColorBorder string
	ColorTitle  string
	Title       string
	Padding     int
}

// NewBox creates a new box with default styling
func NewBox(width int) *Box {
	return &Box{
		Style:       BorderHeavy,
		Width:       width,
		ColorBorder: ColorBorderPrimary,
		ColorTitle:  ColorNeonCyan,
		Padding:     1,
	}
}

// RenderTop renders the top border with optional title
func (b *Box) RenderTop(out io.Writer, l Layout) {
	pad := leftPad(l.HPad())
	innerWidth := b.Width - 2 // Account for left and right borders

	if b.Title == "" {
		// Simple top border without title
		line := b.Style.TopLeft + strings.Repeat(b.Style.Horizontal, innerWidth) + b.Style.TopRight
		fmt.Fprintln(out, pad+style(line, b.ColorBorder))
		return
	}

	// Top border with centered title
	titleText := " " + b.Title + " "
	titleWidth := utf8.RuneCountInString(titleText)
	
	if titleWidth >= innerWidth {
		// Title too wide, just show border
		line := b.Style.TopLeft + strings.Repeat(b.Style.Horizontal, innerWidth) + b.Style.TopRight
		fmt.Fprintln(out, pad+style(line, b.ColorBorder))
		return
	}

	// Calculate padding for centered title
	leftPad := (innerWidth - titleWidth) / 2
	rightPad := innerWidth - titleWidth - leftPad

	var sb strings.Builder
	sb.WriteString(style(b.Style.TopLeft, b.ColorBorder))
	sb.WriteString(style(strings.Repeat(b.Style.Horizontal, leftPad), b.ColorBorder))
	sb.WriteString(style(titleText, b.ColorTitle))
	sb.WriteString(style(strings.Repeat(b.Style.Horizontal, rightPad), b.ColorBorder))
	sb.WriteString(style(b.Style.TopRight, b.ColorBorder))

	fmt.Fprintln(out, pad+sb.String())
}

// RenderBottom renders the bottom border
func (b *Box) RenderBottom(out io.Writer, l Layout) {
	pad := leftPad(l.HPad())
	innerWidth := b.Width - 2
	line := b.Style.BottomLeft + strings.Repeat(b.Style.Horizontal, innerWidth) + b.Style.BottomRight
	fmt.Fprintln(out, pad+style(line, b.ColorBorder))
}

// RenderLine renders a content line with side borders
func (b *Box) RenderLine(out io.Writer, l Layout, content string) {
	pad := leftPad(l.HPad())
	innerWidth := b.Width - 2 - (b.Padding * 2)
	
	// Strip ANSI to calculate visible width
	visibleContent := stripANSI(content)
	visibleWidth := utf8.RuneCountInString(visibleContent)
	
	// Truncate or pad content to fit
	if visibleWidth > innerWidth {
		// Truncate
		rs := []rune(visibleContent)
		if innerWidth > 3 {
			content = string(rs[:innerWidth-3]) + "..."
		} else {
			content = string(rs[:innerWidth])
		}
		visibleWidth = innerWidth
	}
	
	// Build the line
	var sb strings.Builder
	sb.WriteString(style(b.Style.Vertical, b.ColorBorder))
	sb.WriteString(strings.Repeat(" ", b.Padding))
	sb.WriteString(content)
	sb.WriteString(strings.Repeat(" ", innerWidth-visibleWidth+b.Padding))
	sb.WriteString(style(b.Style.Vertical, b.ColorBorder))
	
	fmt.Fprintln(out, pad+sb.String())
}

// RenderEmpty renders an empty line with just borders
func (b *Box) RenderEmpty(out io.Writer, l Layout) {
	b.RenderLine(out, l, "")
}

// RenderDivider renders a horizontal divider within the box
func (b *Box) RenderDivider(out io.Writer, l Layout) {
	pad := leftPad(l.HPad())
	innerWidth := b.Width - 2
	line := b.Style.TeeLeft + strings.Repeat(b.Style.Horizontal, innerWidth) + b.Style.TeeRight
	fmt.Fprintln(out, pad+style(line, b.ColorBorder))
}

// FramedSection creates a complete framed section with content
type FramedSection struct {
	Title   string
	Content []string
	Width   int
}

// RenderFramedSection renders a complete bordered section
func RenderFramedSection(out io.Writer, section FramedSection) {
	l := LayoutFor(out)
	
	// Determine width
	width := section.Width
	if width <= 0 {
		width = l.ContentWidth()
	}
	if width > l.TerminalWidth-l.HPad()*2 {
		width = l.TerminalWidth - l.HPad()*2
	}
	if width < 20 {
		width = 20
	}
	
	box := NewBox(width)
	box.Title = section.Title
	
	// Top border
	box.RenderTop(out, l)
	
	// Empty line after top
	box.RenderEmpty(out, l)
	
	// Content
	for _, line := range section.Content {
		box.RenderLine(out, l, line)
	}
	
	// Empty line before bottom
	box.RenderEmpty(out, l)
	
	// Bottom border
	box.RenderBottom(out, l)
}

// RenderSplashBox renders a border around splash screen content
func RenderSplashBox(out io.Writer, titleLines []string, tagline, author, caption string) {
	l := LayoutFor(out)
	
	// Calculate required width
	maxWidth := 0
	for _, line := range titleLines {
		w := utf8.RuneCountInString(stripANSI(line))
		if w > maxWidth {
			maxWidth = w
		}
	}
	
	// Account for tagline, author, caption
	for _, text := range []string{tagline, author, caption} {
		w := utf8.RuneCountInString(stripANSI(text))
		if w > maxWidth {
			maxWidth = w
		}
	}
	
	// Add padding
	boxWidth := maxWidth + 8
	if boxWidth > l.TerminalWidth-l.HPad()*2 {
		boxWidth = l.TerminalWidth - l.HPad()*2
	}
	if boxWidth < 30 {
		boxWidth = 30
	}
	
	box := NewBox(boxWidth)
	box.Style = BorderDouble
	box.ColorBorder = ColorNeonCyan
	
	// Top border
	box.RenderTop(out, l)
	box.RenderEmpty(out, l)
	
	// Title lines (centered)
	for _, line := range titleLines {
		box.RenderLine(out, l, centerInBox(line, boxWidth-4))
	}
	
	box.RenderEmpty(out, l)
	
	// Tagline
	if strings.TrimSpace(stripANSI(tagline)) != "" {
		box.RenderLine(out, l, centerInBox(tagline, boxWidth-4))
		box.RenderEmpty(out, l)
	}
	
	// Author
	if strings.TrimSpace(stripANSI(author)) != "" {
		box.RenderLine(out, l, centerInBox(author, boxWidth-4))
	}
	
	// Caption
	if strings.TrimSpace(stripANSI(caption)) != "" {
		box.RenderEmpty(out, l)
		box.RenderLine(out, l, centerInBox(caption, boxWidth-4))
	}
	
	box.RenderEmpty(out, l)
	
	// Bottom border
	box.RenderBottom(out, l)
}

// centerInBox centers text within a given width
func centerInBox(text string, width int) string {
	visible := stripANSI(text)
	visibleWidth := utf8.RuneCountInString(visible)
	
	if visibleWidth >= width {
		return text
	}
	
	leftPad := (width - visibleWidth) / 2
	return strings.Repeat(" ", leftPad) + text
}
