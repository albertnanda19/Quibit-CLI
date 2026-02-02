package tui

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

// ShowSplashScreen renders a startup splash once per process (caller enforces once).
// It is terminal-safe: if stdin/stdout are not TTY, it becomes a no-op.
func ShowSplashScreen(in *os.File, out io.Writer) error {
	if in == nil || out == nil {
		return nil
	}
	if !isTerminal(int(in.Fd())) {
		return nil
	}
	type fdWriter interface{ Fd() uintptr }
	fw, ok := out.(fdWriter)
	if !ok || !isTerminal(int(fw.Fd())) {
		return nil
	}

	// Clear screen and move cursor home.
	fmt.Fprint(out, "\033[2J\033[H")

	// Render spaced title: Q U I B I T (block-style, symmetric, calm).
	q := []string{
		" █████ ",
		"██   ██",
		"██   ██",
		"██ █ ██",
		"██  ███",
		" ██████",
	}
	u := []string{
		"██   ██",
		"██   ██",
		"██   ██",
		"██   ██",
		"██   ██",
		" █████ ",
	}
	i := []string{
		"██████",
		"  ██  ",
		"  ██  ",
		"  ██  ",
		"  ██  ",
		"██████",
	}
	b := []string{
		"█████ ",
		"██  ██",
		"█████ ",
		"██  ██",
		"██  ██",
		"█████ ",
	}
	t := []string{
		"██████",
		"  ██  ",
		"  ██  ",
		"  ██  ",
		"  ██  ",
		"  ██  ",
	}

	BlankLine(out)
	for row := 0; row < 6; row++ {
		line := q[row] + "  " + u[row] + "  " + i[row] + "  " + b[row] + "  " + i[row] + "  " + t[row]
		fmt.Fprintln(out, style(line, ColorStatus))
	}
	BlankLine(out)
	fmt.Fprintln(out, style("Design. Generate. Iterate.", ColorMuted))
	fmt.Fprintln(out, style("by Albert Mangiri", ColorGroupHeader))
	BlankLine(out)
	fmt.Fprintln(out, style("Press Enter to continue", ColorMuted))

	r := bufio.NewReader(in)
	for {
		b, err := r.ReadByte()
		if err != nil {
			return nil
		}
		if b == '\n' || b == '\r' {
			break
		}
	}

	// Keep the transition clean for the existing flow.
	fmt.Fprint(out, "\033[2J\033[H")
	return nil
}

