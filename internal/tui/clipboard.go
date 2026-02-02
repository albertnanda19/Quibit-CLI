package tui

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"os/exec"

	"golang.org/x/sys/unix"
)

// CopyToClipboard tries to copy text to the system clipboard.
// Strategy:
// 1) External clipboard tools (wl-copy/xclip/xsel/pbcopy)
// 2) OSC52 escape sequence (terminal-dependent)
func CopyToClipboard(out io.Writer, text string) error {
	if err := copyWithExternalTools(text); err == nil {
		return nil
	}
	if err := copyWithOSC52(out, text); err == nil {
		return nil
	}
	return fmt.Errorf("clipboard is unavailable (install wl-copy/xclip/xsel, or use a terminal that supports OSC52)")
}

func copyWithExternalTools(text string) error {
	// Wayland
	if _, err := exec.LookPath("wl-copy"); err == nil {
		cmd := exec.Command("wl-copy")
		cmd.Stdin = bytes.NewReader([]byte(text))
		return cmd.Run()
	}

	// X11
	if _, err := exec.LookPath("xclip"); err == nil {
		cmd := exec.Command("xclip", "-selection", "clipboard")
		cmd.Stdin = bytes.NewReader([]byte(text))
		return cmd.Run()
	}
	if _, err := exec.LookPath("xsel"); err == nil {
		cmd := exec.Command("xsel", "--clipboard", "--input")
		cmd.Stdin = bytes.NewReader([]byte(text))
		return cmd.Run()
	}

	// macOS
	if _, err := exec.LookPath("pbcopy"); err == nil {
		cmd := exec.Command("pbcopy")
		cmd.Stdin = bytes.NewReader([]byte(text))
		return cmd.Run()
	}

	return fmt.Errorf("no external clipboard tool found")
}

func copyWithOSC52(out io.Writer, text string) error {
	// Prefer writing OSC52 directly to the real TTY, since Cobra/out can be wrapped.
	if tty, err := os.OpenFile("/dev/tty", os.O_WRONLY, 0); err == nil {
		defer func() { _ = tty.Close() }()
		return writeOSC52(tty, text)
	}

	// Fallback: try provided writer if it's a terminal, else try stdout.
	if f, ok := out.(*os.File); ok && isTerminalFD(int(f.Fd())) {
		return writeOSC52(f, text)
	}
	if isTerminalFD(int(os.Stdout.Fd())) {
		return writeOSC52(os.Stdout, text)
	}

	// OSC52 payload is often size-limited; keep it conservative.
	const maxBytes = 100_000
	b := []byte(text)
	if len(b) > maxBytes {
		return fmt.Errorf("content is too large to copy via terminal clipboard")
	}
	return fmt.Errorf("no compatible terminal available for OSC52 clipboard")
}

func isTerminalFD(fd int) bool {
	_, err := unix.IoctlGetTermios(fd, unix.TCGETS)
	return err == nil
}

func writeOSC52(f *os.File, text string) error {
	if f == nil {
		return fmt.Errorf("terminal is unavailable")
	}
	enc := base64.StdEncoding.EncodeToString([]byte(text))
	// OSC52: ESC ] 52 ; c ; <base64> BEL
	_, err := fmt.Fprintf(f, "\033]52;c;%s\a", enc)
	return err
}
