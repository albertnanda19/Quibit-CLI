package tui

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/sys/unix"
)

type Option struct {
	ID    string
	Label string
}

type SelectEntry struct {
	ID         string
	Label      string
	Selectable bool
}

func SelectOption(in *os.File, out io.Writer, prompt string, options []Option) (Option, error) {
	return SelectOptionWithDefault(in, out, prompt, options, "")
}

func SelectOptionWithDefault(in *os.File, out io.Writer, prompt string, options []Option, defaultID string) (Option, error) {
	entries := make([]SelectEntry, 0, len(options))
	for i := range options {
		entries = append(entries, SelectEntry{
			ID:         options[i].ID,
			Label:      options[i].Label,
			Selectable: true,
		})
	}
	selection, err := SelectEntriesWithDefault(in, out, prompt, entries, defaultID)
	if err != nil {
		return Option{}, err
	}
	return Option{ID: selection.ID, Label: selection.Label}, nil
}

func SelectEntries(in *os.File, out io.Writer, prompt string, entries []SelectEntry) (SelectEntry, error) {
	return SelectEntriesWithDefault(in, out, prompt, entries, "")
}

func SelectEntriesWithDefault(in *os.File, out io.Writer, prompt string, entries []SelectEntry, defaultID string) (SelectEntry, error) {
	if len(entries) == 0 {
		return SelectEntry{}, errors.New("no options available")
	}

	fd := int(in.Fd())
	if !isTerminal(fd) {
		return SelectEntry{}, errors.New("interactive selection requires a terminal (TTY)")
	}

	restore, err := makeRaw(fd)
	if err != nil {
		return SelectEntry{}, fmt.Errorf("unable to enable interactive mode: %w", err)
	}
	defer restore()

	if prompt != "" {
		BlankLine(out)
		Context(out, prompt)
		Divider(out)
		BlankLine(out)
	}
	selected := findDefaultSelectableIndex(entries, defaultID)
	renderEntries(out, entries, selected)

	for {
		key, err := readKey(in)
		if err != nil {
			return SelectEntry{}, err
		}

		switch key {
		case keyUp:
			selected = moveSelectable(entries, selected, -1)
		case keyDown:
			selected = moveSelectable(entries, selected, +1)
		case keyEnter:
			if selected >= 0 && selected < len(entries) && entries[selected].Selectable {
				fmt.Fprintln(out, "")
				return entries[selected], nil
			}
			continue
		default:
			continue
		}

		moveCursorUp(out, len(entries)+selectFooterLines)
		renderEntries(out, entries, selected)
	}
}

func findDefaultSelectableIndex(entries []SelectEntry, defaultID string) int {
	if defaultID != "" {
		for i := range entries {
			if entries[i].Selectable && entries[i].ID == defaultID {
				return i
			}
		}
	}
	for i := range entries {
		if entries[i].Selectable {
			return i
		}
	}
	return 0
}

func moveSelectable(entries []SelectEntry, selected int, dir int) int {
	if len(entries) == 0 {
		return 0
	}
	if dir == 0 {
		return selected
	}
	next := selected
	for {
		candidate := next + dir
		if candidate < 0 || candidate >= len(entries) {
			return next
		}
		next = candidate
		if entries[next].Selectable {
			return next
		}
	}
}

const (
	keyUnknown = iota
	keyUp
	keyDown
	keyEnter
)

func readKey(in *os.File) (int, error) {
	b, err := readByte(in)
	if err != nil {
		return keyUnknown, err
	}
	switch b {
	case '\r', '\n':
		return keyEnter, nil
	case 0x1b:
		b2, err := readByte(in)
		if err != nil {
			return keyUnknown, err
		}
		if b2 != '[' {
			return keyUnknown, nil
		}
		b3, err := readByte(in)
		if err != nil {
			return keyUnknown, err
		}
		switch b3 {
		case 'A':
			return keyUp, nil
		case 'B':
			return keyDown, nil
		default:
			return keyUnknown, nil
		}
	default:
		return keyUnknown, nil
	}
}

func readByte(in *os.File) (byte, error) {
	var buf [1]byte
	_, err := in.Read(buf[:])
	if err != nil {
		return 0, fmt.Errorf("read input: %w", err)
	}
	return buf[0], nil
}

const selectFooterLines = 2

func renderEntries(out io.Writer, entries []SelectEntry, selected int) {
	l := LayoutFor(out)
	width := l.ContentWidth()
	pad := leftPad(l.HPad())
	for i := range entries {
		label := sanitizeOneLine(entries[i].Label)
		label = truncateToWidth(label, width-6) // room for selection prefix + spacing
		if !entries[i].Selectable {
			fmt.Fprintf(out, "\r\033[K%s%s\n", pad, style(label, ColorGroupHeader))
			continue
		}
		if i == selected {
			prefix := style("› ", ColorAccent)
			fmt.Fprintf(out, "\r\033[K%s%s%s\n", pad, prefix, style(label, ColorPrimary))
			continue
		}
		fmt.Fprintf(out, "\r\033[K%s  %s\n", pad, style(label, ColorBody))
	}

	fmt.Fprint(out, "\r\033[K\n")
	fmt.Fprintf(out, "\r\033[K%s%s\n", pad, style("↑/↓ navigate  ·  Enter select", ColorMuted))
}

func moveCursorUp(out io.Writer, lines int) {
	if lines <= 0 {
		return
	}
	fmt.Fprintf(out, "\033[%dA", lines)
}

func terminalWidth(out io.Writer) int {

	const fallback = 100

	type fdWriter interface{ Fd() uintptr }
	f, ok := out.(fdWriter)
	if !ok {
		return fallback
	}
	ws, err := unix.IoctlGetWinsize(int(f.Fd()), unix.TIOCGWINSZ)
	if err != nil || ws == nil || ws.Col == 0 {
		return fallback
	}
	return int(ws.Col)
}

func sanitizeOneLine(s string) string {
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}

func truncateToWidth(s string, width int) string {
	if width <= 0 {
		return ""
	}

	if width < 10 {
		width = 10
	}
	rs := []rune(s)
	if len(rs) <= width {
		return s
	}
	if width <= 3 {
		return string(rs[:width])
	}
	return string(rs[:width-3]) + "..."
}

func isTerminal(fd int) bool {
	_, err := unix.IoctlGetTermios(fd, unix.TCGETS)
	return err == nil
}

func makeRaw(fd int) (func(), error) {
	oldState, err := unix.IoctlGetTermios(fd, unix.TCGETS)
	if err != nil {
		return nil, err
	}
	newState := *oldState
	newState.Iflag &^= unix.IGNBRK | unix.BRKINT | unix.PARMRK | unix.ISTRIP | unix.INLCR | unix.IGNCR | unix.ICRNL | unix.IXON
	newState.Oflag &^= unix.OPOST
	newState.Lflag &^= unix.ECHO | unix.ECHONL | unix.ICANON | unix.ISIG | unix.IEXTEN
	newState.Cflag &^= unix.CSIZE | unix.PARENB
	newState.Cflag |= unix.CS8
	newState.Cc[unix.VMIN] = 1
	newState.Cc[unix.VTIME] = 0

	if err := unix.IoctlSetTermios(fd, unix.TCSETS, &newState); err != nil {
		return nil, err
	}
	restore := func() {
		_ = unix.IoctlSetTermios(fd, unix.TCSETS, oldState)
	}
	return restore, nil
}
