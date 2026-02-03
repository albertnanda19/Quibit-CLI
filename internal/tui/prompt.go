package tui

import "io"

func PromptPrefix(out io.Writer) string {
	l := LayoutFor(out)
	return leftPad(l.HPad()) + "> "
}
