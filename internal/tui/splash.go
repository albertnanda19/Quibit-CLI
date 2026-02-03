package tui

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"strings"
	"time"
	"unicode/utf8"
)

func ShowSplashScreen(ctx context.Context, in *os.File, out io.Writer, mode string) (shown bool, err error) {
	if in == nil || out == nil {
		return false, nil
	}
	if !isTerminal(int(in.Fd())) {
		return false, nil
	}
	type fdWriter interface{ Fd() uintptr }
	fw, ok := out.(fdWriter)
	if !ok || !isTerminal(int(fw.Fd())) {
		return false, nil
	}

	version := buildVersion()
	l := LayoutFor(out)

	titleLines := splashHeroTitleLines(l.ContentWidth())
	tagline := style(splashTagline(mode, l.ContentWidth()), ColorMuted)
	author := style("by Albert Mangiri", ColorGroupHeader)
	caption := style(splashCaption(mode, version), ColorDivider)

	blocks := [][]string{
		alignBlock(titleLines, l, true),
		{""},
		alignBlock([]string{tagline}, l, true),
		{""},
		alignBlock([]string{author}, l, true),
	}
	if strings.TrimSpace(stripANSI(caption)) != "" {
		blocks = append(blocks, alignBlock([]string{caption}, l, true))
	}

	revealBlocks(ctx, out, blocks)
	fmt.Fprintln(out, "")
	return true, nil
}

func splashTagline(mode string, width int) string {
	mode = strings.ToLower(strings.TrimSpace(mode))
	if mode == "" {
		mode = "idle"
	}

	var base string
	switch mode {
	case "generate":
		base = "Shaping a new project."
	case "continue":
		base = "Evolving an existing idea."
	case "browse":
		base = "Reviewing your work."
	default:
		base = "Engineering ideas, intentionally."
		mode = "idle"
	}

	if mode != "idle" && width >= 44 {
		base = base + style("  ·  "+mode, ColorDivider)
	}
	return base
}

func splashCaption(mode string, version string) string {
	mode = strings.ToLower(strings.TrimSpace(mode))
	version = strings.TrimSpace(version)
	parts := []string{}
	if version != "" {
		parts = append(parts, version)
	}
	if mode != "" && mode != "idle" {
		parts = append(parts, mode)
	}
	return strings.Join(parts, "  ·  ")
}

func splashHeroTitleLines(width int) []string {
	// Bold, block-style hero title. Keep stable across modes.
	// Pick a compact fallback if the terminal is too narrow (never clip).

	// 5-line, ~47 cols (including spaces)
	bigLetters := map[string][]string{
		"Q": {
			" █████ ",
			"█     █",
			"█     █",
			"█   █ █",
			" ██████",
		},
		"U": {
			"█     █",
			"█     █",
			"█     █",
			"█     █",
			" █████ ",
		},
		"I": {
			"███████",
			"   █   ",
			"   █   ",
			"   █   ",
			"███████",
		},
		"B": {
			"██████ ",
			"█     █",
			"██████ ",
			"█     █",
			"██████ ",
		},
		"T": {
			"███████",
			"   █   ",
			"   █   ",
			"   █   ",
			"   █   ",
		},
	}
	orderBig := []string{"Q", "U", "I", "B", "I", "T"}
	big := make([]string, 0, 5)
	for row := 0; row < 5; row++ {
		parts := make([]string, 0, len(orderBig))
		for _, k := range orderBig {
			parts = append(parts, bigLetters[k][row])
		}
		big = append(big, strings.Join(parts, " "))
	}
	big3D := extrudeBlockStyled(big, 2, 1)

	// 3-line, ~29 cols
	compactLetters := map[string][]string{
		"Q": {"███▄", "█ ██", "███▄"},
		"U": {"█  █", "█  █", "████"},
		"I": {"████", " ██ ", "████"},
		"B": {"███▄", "███▄", "███▄"},
		"T": {"████", " ██ ", " ██ "},
	}
	orderCompact := []string{"Q", "U", "I", "B", "I", "T"}
	compact := make([]string, 0, 3)
	for row := 0; row < 3; row++ {
		parts := make([]string, 0, len(orderCompact))
		for _, k := range orderCompact {
			parts = append(parts, compactLetters[k][row])
		}
		compact = append(compact, strings.Join(parts, " "))
	}

	if width <= 0 {
		return styleBlock(compact, ColorPrimary)
	}
	if blockWidth(big3D) <= width {
		return big3D
	}
	if blockWidth(big) <= width {
		return styleBlock(big, ColorPrimary)
	}
	if blockWidth(compact) <= width {
		return styleBlock(compact, ColorPrimary)
	}
	// Extreme narrow fallback: still readable, never clipped.
	return []string{style("QUIBIT", ColorPrimary)}
}

func styleBlock(lines []string, sgr string) []string {
	out := make([]string, 0, len(lines))
	for _, ln := range lines {
		out = append(out, style(ln, sgr))
	}
	return out
}

func extrudeBlockStyled(lines []string, dx int, dy int) []string {
	if len(lines) == 0 {
		return nil
	}
	if dx < 0 {
		dx = 0
	}
	if dy < 0 {
		dy = 0
	}

	base := make([][]rune, 0, len(lines))
	maxW := 0
	for _, ln := range lines {
		rs := []rune(strings.TrimRight(ln, "\r\n"))
		base = append(base, rs)
		if len(rs) > maxW {
			maxW = len(rs)
		}
	}
	h := len(base)
	w := maxW

	ch := h + dy
	cw := w + dx
	const (
		cellSpace  = 0
		cellShadow = 1
		cellFace   = 2
	)
	typeMap := make([][]uint8, ch)
	for y := 0; y < ch; y++ {
		row := make([]uint8, cw)
		for x := 0; x < cw; x++ {
			row[x] = cellSpace
		}
		typeMap[y] = row
	}

	filled := func(y int, x int) bool {
		if y < 0 || y >= h {
			return false
		}
		if x < 0 || x >= len(base[y]) {
			return false
		}
		return base[y][x] != ' '
	}
	markShadow := func(y int, x int) {
		cy := y + dy
		cx := x + dx
		if cy < 0 || cy >= ch || cx < 0 || cx >= cw {
			return
		}
		if typeMap[cy][cx] == cellSpace {
			typeMap[cy][cx] = cellShadow
		}
		// Thicken subtly for dx>1 so the extrusion reads as a plane.
		if dx > 1 {
			if cx-1 >= 0 && typeMap[cy][cx-1] == cellSpace {
				typeMap[cy][cx-1] = cellShadow
			}
		}
	}
	for y := 0; y < h; y++ {
		for x := 0; x < len(base[y]); x++ {
			if !filled(y, x) {
				continue
			}
			// Only cast shadow from edges (right/bottom) to avoid noisy interior fills.
			rightEdge := !filled(y, x+1)
			bottomEdge := !filled(y+1, x)
			if rightEdge || bottomEdge {
				markShadow(y, x)
			}
		}
	}
	for y := 0; y < h; y++ {
		for x := 0; x < len(base[y]); x++ {
			if base[y][x] == ' ' {
				continue
			}
			if y >= 0 && y < ch && x >= 0 && x < cw {
				typeMap[y][x] = cellFace
			}
		}
	}

	shadowRune := '█'
	if noColor() {
		shadowRune = '▓'
	}

	faceOn := ansi(ColorPrimary)
	shadowOn := ansi(ColorDivider)
	reset := ansiReset
	if noColor() {
		faceOn = ""
		shadowOn = ""
		reset = ""
	}

	out := make([]string, 0, ch)
	for y := 0; y < ch; y++ {
		var b strings.Builder
		cur := uint8(255)
		writeMode := func(mode uint8) {
			if noColor() {
				return
			}
			if mode == cur {
				return
			}
			switch mode {
			case cellFace:
				b.WriteString(faceOn)
			case cellShadow:
				b.WriteString(shadowOn)
			default:
				b.WriteString(reset)
			}
			cur = mode
		}

		for x := 0; x < cw; x++ {
			mode := typeMap[y][x]
			switch mode {
			case cellFace:
				writeMode(cellFace)
				b.WriteRune('█')
			case cellShadow:
				writeMode(cellShadow)
				b.WriteRune(shadowRune)
			default:
				writeMode(cellSpace)
				b.WriteRune(' ')
			}
		}
		if !noColor() {
			b.WriteString(reset)
		}
		ln := strings.TrimRight(b.String(), " \t")
		out = append(out, ln)
	}
	for len(out) > 0 && strings.TrimSpace(stripANSI(out[len(out)-1])) == "" {
		out = out[:len(out)-1]
	}
	return out
}

func blockWidth(lines []string) int {
	max := 0
	for _, ln := range lines {
		ln = strings.TrimRight(ln, " \t\r\n")
		n := utf8.RuneCountInString(stripANSI(ln))
		if n > max {
			max = n
		}
	}
	return max
}

func alignBlock(lines []string, l Layout, allowCenter bool) []string {
	padLeft := leftPad(l.HPad())
	if !allowCenter || !l.CenterHeaders() {
		out := make([]string, 0, len(lines))
		for _, ln := range lines {
			out = append(out, padLeft+strings.TrimRight(ln, "\r\n"))
		}
		return out
	}

	w := l.ContentWidth()
	out := make([]string, 0, len(lines))
	for _, ln := range lines {
		ln = strings.TrimRight(ln, "\r\n")
		visible := strings.TrimRight(ln, " \t")
		n := utf8.RuneCountInString(stripANSI(visible))
		if w <= 0 || n <= 0 || n >= w {
			out = append(out, padLeft+ln)
			continue
		}
		pad := (w - n) / 2
		out = append(out, leftPad(l.HPad()+pad)+ln)
	}
	return out
}

func revealLines(ctx context.Context, out io.Writer, lines []string) {
	if len(lines) == 0 {
		return
	}
	if !motionAllowed(out) {
		for _, line := range lines {
			fmt.Fprintln(out, line)
		}
		return
	}

	const perLine = 70 * time.Millisecond
	for i := range lines {
		fmt.Fprintln(out, lines[i])

		if i == len(lines)-1 {
			break
		}
		timer := time.NewTimer(perLine)
		select {
		case <-ctx.Done():
			timer.Stop()

			for j := i + 1; j < len(lines); j++ {
				fmt.Fprintln(out, lines[j])
			}
			return
		case <-timer.C:
		}
	}
}

func revealBlocks(ctx context.Context, out io.Writer, blocks [][]string) {
	if len(blocks) == 0 {
		return
	}
	if !motionAllowed(out) {
		for _, b := range blocks {
			for _, line := range b {
				fmt.Fprintln(out, line)
			}
		}
		return
	}

	// Reveal per block (presence), not per character (performance theatre).
	// Keep total duration within ~300–600ms.
	perBlock := 0 * time.Millisecond
	if len(blocks) > 1 {
		perBlock = (480 * time.Millisecond) / time.Duration(len(blocks)-1)
		if perBlock < 80*time.Millisecond {
			perBlock = 80 * time.Millisecond
		}
		if perBlock > 140*time.Millisecond {
			perBlock = 140 * time.Millisecond
		}
	}
	for i := range blocks {
		for _, line := range blocks[i] {
			fmt.Fprintln(out, line)
		}
		if i == len(blocks)-1 {
			break
		}
		timer := time.NewTimer(perBlock)
		select {
		case <-ctx.Done():
			timer.Stop()
			// Flush remaining blocks immediately.
			for j := i + 1; j < len(blocks); j++ {
				for _, line := range blocks[j] {
					fmt.Fprintln(out, line)
				}
			}
			return
		case <-timer.C:
		}
	}
}

func centerLine(s string, width int) string {
	s = strings.TrimRight(s, "\r\n")
	if width <= 0 {
		return s
	}
	n := utf8.RuneCountInString(stripANSI(s))
	if n <= 0 || n >= width {
		return s
	}
	pad := (width - n) / 2
	if pad <= 0 {
		return s
	}
	return strings.Repeat(" ", pad) + s
}

func buildVersion() string {

	bi, ok := debug.ReadBuildInfo()
	if !ok || bi == nil {
		return ""
	}
	v := strings.TrimSpace(bi.Main.Version)
	if v == "" || v == "(devel)" {
		return ""
	}

	const max = 24
	if utf8.RuneCountInString(v) > max {
		rs := []rune(v)
		v = string(rs[:max-1]) + "…"
	}
	return v
}
