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

func SelectOption(in *os.File, out io.Writer, prompt string, options []Option) (Option, error) {
	return SelectOptionWithDefault(in, out, prompt, options, "")
}

func SelectOptionWithDefault(in *os.File, out io.Writer, prompt string, options []Option, defaultID string) (Option, error) {
	if len(options) == 0 {
		return Option{}, errors.New("no options available")
	}

	fd := int(in.Fd())
	if !isTerminal(fd) {
		return Option{}, errors.New("interactive selection requires a terminal (TTY)")
	}

	restore, err := makeRaw(fd)
	if err != nil {
		return Option{}, fmt.Errorf("unable to enable interactive mode: %w", err)
	}
	defer restore()

	if prompt != "" {
		BlankLine(out)
		fmt.Fprintln(out, prompt)
		ControlsSelect(out)
		BlankLine(out)
	}
	selected := findDefaultIndex(options, defaultID)
	renderOptions(out, options, selected)

	for {
		key, err := readKey(in)
		if err != nil {
			return Option{}, err
		}

		switch key {
		case keyUp:
			if selected > 0 {
				selected--
			}
		case keyDown:
			if selected < len(options)-1 {
				selected++
			}
		case keyEnter:
			fmt.Fprintln(out, "")
			return options[selected], nil
		default:
			continue
		}

		moveCursorUp(out, len(options))
		renderOptions(out, options, selected)
	}
}

func findDefaultIndex(options []Option, defaultID string) int {
	if defaultID == "" {
		return 0
	}
	for i := range options {
		if options[i].ID == defaultID {
			return i
		}
	}
	return 0
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

func renderOptions(out io.Writer, options []Option, selected int) {
	width := terminalWidth(out)
	for i := range options {
		prefix := "  "
		if i == selected {
			prefix = "> "
		}
		label := sanitizeOneLine(options[i].Label)
		label = truncateToWidth(label, width-len(prefix))
		fmt.Fprintf(out, "\r\033[K%s%s\n", prefix, label)
	}
}

func moveCursorUp(out io.Writer, lines int) {
	if lines <= 0 {
		return
	}
	fmt.Fprintf(out, "\033[%dA", lines)
}

func terminalWidth(out io.Writer) int {
	// Default to a safe width if we can't detect terminal size.
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
	// avoid weird tiny widths
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
