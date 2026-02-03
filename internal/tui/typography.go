package tui

import "unicode/utf8"

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

