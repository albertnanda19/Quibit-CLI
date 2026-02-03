package tui

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Motion/animation helpers.
//
// Design goals:
// - subtle + calm (low frequency)
// - safe to disable globally
// - spinner never blocks caller goroutine
// - no flow/business logic changes (UI only)

var motionEnabled atomic.Bool

func init() {
	// Default: enabled, unless explicitly disabled by env.
	motionEnabled.Store(true)
}

// SetMotionEnabled toggles terminal micro-animations (spinner, micro-delays).
// This is intended to be wired from a single CLI flag (e.g. --no-anim).
func SetMotionEnabled(v bool) { motionEnabled.Store(v) }

func motionDisabledByEnv() bool {
	// Single environment kill-switch (optional).
	if v := strings.TrimSpace(os.Getenv("QUIBIT_NO_ANIM")); v != "" && v != "0" && strings.ToLower(v) != "false" {
		return true
	}
	return false
}

func motionAllowed(out io.Writer) bool {
	if out == nil {
		return false
	}
	if motionDisabledByEnv() {
		return false
	}
	if !motionEnabled.Load() {
		return false
	}
	// Only animate if stdout is a TTY.
	type fdWriter interface{ Fd() uintptr }
	fw, ok := out.(fdWriter)
	if !ok {
		return false
	}
	return isTerminal(int(fw.Fd()))
}

// Transition is a tiny pause used between screens/modes to avoid a harsh jump.
// Keep it subtle; never use this for per-line effects.
//
// It is cancellable via ctx so it can be interrupted immediately on shutdown.
func Transition(ctx context.Context, out io.Writer) {
	if !motionAllowed(out) {
		return
	}
	const d = 90 * time.Millisecond
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return
	case <-timer.C:
		return
	}
}

// Spinner is a subtle, low-frequency loading indicator rendered on a single line.
// It is always stoppable (Stop) and runs on its own goroutine.
type Spinner struct {
	out     io.Writer
	message string

	stopOnce sync.Once
	stopCh   chan struct{}
	doneCh   chan struct{}

	mu sync.Mutex // serialize spinner writes vs Stop() clear
}

// StartSpinner begins rendering a subtle spinner until Stop() is called or ctx is done.
// When motion is disabled or output is not a TTY, it becomes a no-op spinner.
func StartSpinner(ctx context.Context, out io.Writer, message string) *Spinner {
	s := &Spinner{
		out:     out,
		message: strings.TrimSpace(message),
		stopCh:  make(chan struct{}),
		doneCh:  make(chan struct{}),
	}
	if s.message == "" {
		s.message = "Working"
	}
	if !motionAllowed(out) {
		close(s.doneCh)
		return s
	}
	go s.loop(ctx)
	return s
}

func (s *Spinner) loop(ctx context.Context) {
	defer close(s.doneCh)

	// Calm, subtle frames; low frequency to avoid “busy” feel.
	frames := []string{"·  ", "·· ", "···"}
	ticker := time.NewTicker(240 * time.Millisecond)
	defer ticker.Stop()

	i := 0
	// Initial paint.
	s.paint(frames[i%len(frames)])

	for {
		select {
		case <-ctx.Done():
			s.clear()
			return
		case <-s.stopCh:
			s.clear()
			return
		case <-ticker.C:
			i++
			s.paint(frames[i%len(frames)])
		}
	}
}

func (s *Spinner) paint(frame string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Single-line redraw: CR + clear line.
	fmt.Fprintf(s.out, "\r\033[K%s", style("• "+s.message+" "+frame, ColorStatus))
}

func (s *Spinner) clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Clear spinner line and end cleanly (so next output starts on a fresh line).
	fmt.Fprint(s.out, "\r\033[K\n")
}

// Stop stops the spinner and clears its line. Safe to call multiple times.
func (s *Spinner) Stop() {
	if s == nil {
		return
	}
	s.stopOnce.Do(func() { close(s.stopCh) })
	<-s.doneCh
}

// Typing is a limited, fast typing micro-effect for 1–2 short lines (headers/status).
// It runs on its own goroutine and can be stopped.
type Typing struct {
	out io.Writer

	stopOnce sync.Once
	stopCh   chan struct{}
	doneCh   chan struct{}

	mu sync.Mutex
}

// StartTypeLine types a single line quickly (non-blocking). If motion is disabled,
// it prints the full line immediately.
func StartTypeLine(ctx context.Context, out io.Writer, line string) *Typing {
	t := &Typing{
		out:    out,
		stopCh: make(chan struct{}),
		doneCh: make(chan struct{}),
	}
	line = strings.TrimRight(line, "\r\n")
	if strings.TrimSpace(line) == "" {
		close(t.doneCh)
		return t
	}
	if !motionAllowed(out) {
		fmt.Fprintln(out, line)
		close(t.doneCh)
		return t
	}
	go t.loop(ctx, line)
	return t
}

func (t *Typing) loop(ctx context.Context, line string) {
	defer close(t.doneCh)
	// Fast, subtle; do not use for long content.
	const perRune = 8 * time.Millisecond
	rs := []rune(line)

	t.mu.Lock()
	fmt.Fprint(t.out, "\r\033[K")
	t.mu.Unlock()

	for i := 0; i < len(rs); i++ {
		select {
		case <-ctx.Done():
			t.flush(line)
			return
		case <-t.stopCh:
			t.flush(line)
			return
		default:
		}

		t.mu.Lock()
		fmt.Fprint(t.out, string(rs[i]))
		t.mu.Unlock()

		timer := time.NewTimer(perRune)
		select {
		case <-ctx.Done():
			timer.Stop()
			t.flush(line)
			return
		case <-t.stopCh:
			timer.Stop()
			t.flush(line)
			return
		case <-timer.C:
		}
	}

	t.mu.Lock()
	fmt.Fprint(t.out, "\n")
	t.mu.Unlock()
}

func (t *Typing) flush(line string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	fmt.Fprintf(t.out, "\r\033[K%s\n", line)
}

// Stop stops the typing effect. Safe to call multiple times.
func (t *Typing) Stop() {
	if t == nil {
		return
	}
	t.stopOnce.Do(func() { close(t.stopCh) })
	<-t.doneCh
}

