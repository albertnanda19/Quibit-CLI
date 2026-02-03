package tui

import (
	"io"
	"strings"
	"unicode/utf8"
)

type WidthCategory string

const (
	WidthNarrow   WidthCategory = "narrow"
	WidthStandard WidthCategory = "standard"
	WidthWide     WidthCategory = "wide"
)

type SpaceToken string

const (
	SpaceSmall  SpaceToken = "small"
	SpaceMedium SpaceToken = "medium"
	SpaceLarge  SpaceToken = "large"
)

type Layout struct {
	TerminalWidth int
	Category      WidthCategory

	HPadding SpaceToken
	VSection SpaceToken

	MaxLine int
}

func LayoutFor(out io.Writer) Layout {
	w := terminalWidth(out)
	if w <= 0 {
		w = 100
	}

	cat := WidthStandard
	switch {
	case w <= 80:
		cat = WidthNarrow
	case w <= 120:
		cat = WidthStandard
	default:
		cat = WidthWide
	}

	h := SpaceMedium
	v := SpaceMedium
	maxLine := 96

	switch cat {
	case WidthNarrow:
		h = SpaceSmall
		v = SpaceSmall
		maxLine = 88
	case WidthStandard:
		h = SpaceMedium
		v = SpaceMedium
		maxLine = 96
	case WidthWide:
		h = SpaceLarge
		v = SpaceMedium
		maxLine = 100
	default:
	}

	return Layout{
		TerminalWidth: w,
		Category:      cat,
		HPadding:      h,
		VSection:      v,
		MaxLine:       maxLine,
	}
}

func (l Layout) HPad() int {
	switch l.HPadding {
	case SpaceSmall:
		return 2
	case SpaceMedium:
		return 4
	case SpaceLarge:
		return 6
	default:
		return 4
	}
}

func (l Layout) VSectionLines() int {
	switch l.VSection {
	case SpaceSmall:
		return 1
	case SpaceMedium:
		return 1
	case SpaceLarge:
		return 2
	default:
		return 1
	}
}

func (l Layout) ContentWidth() int {
	if l.TerminalWidth <= 0 {
		return 0
	}

	availAfterLeftPad := l.TerminalWidth - l.HPad()
	if availAfterLeftPad <= 0 {
		return 0
	}

	w := l.TerminalWidth - 2*l.HPad()
	if w <= 0 {
		w = availAfterLeftPad
	}
	if w > l.MaxLine {
		w = l.MaxLine
	}
	if w > availAfterLeftPad {
		w = availAfterLeftPad
	}
	if w < 10 {
		w = 10
		if w > availAfterLeftPad {
			w = availAfterLeftPad
		}
	}
	return w
}

func (l Layout) CenterHeaders() bool {

	return l.Category != WidthNarrow
}

func leftPad(n int) string {
	if n <= 0 {
		return ""
	}
	return strings.Repeat(" ", n)
}

func visibleRuneLen(s string) int {
	return utf8.RuneCountInString(stripANSI(s))
}
