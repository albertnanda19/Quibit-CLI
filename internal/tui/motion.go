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

var motionEnabled atomic.Bool

func init() {

	motionEnabled.Store(true)
}

func SetMotionEnabled(v bool) { motionEnabled.Store(v) }

func motionDisabledByEnv() bool {

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

	type fdWriter interface{ Fd() uintptr }
	fw, ok := out.(fdWriter)
	if !ok {
		return false
	}
	return isTerminal(int(fw.Fd()))
}

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

type Spinner struct {
	out     io.Writer
	message string

	stopOnce sync.Once
	stopCh   chan struct{}
	doneCh   chan struct{}

	mu sync.Mutex
}

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

	frames := []string{"·  ", "·· ", "···"}
	ticker := time.NewTicker(240 * time.Millisecond)
	defer ticker.Stop()

	i := 0

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

	fmt.Fprintf(s.out, "\r\033[K%s", style("• "+s.message+" "+frame, ColorStatus))
}

func (s *Spinner) clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	fmt.Fprint(s.out, "\r\033[K\n")
}

func (s *Spinner) Stop() {
	if s == nil {
		return
	}
	s.stopOnce.Do(func() { close(s.stopCh) })
	<-s.doneCh
}

type Typing struct {
	out io.Writer

	stopOnce sync.Once
	stopCh   chan struct{}
	doneCh   chan struct{}

	mu sync.Mutex
}

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

func (t *Typing) Stop() {
	if t == nil {
		return
	}
	t.stopOnce.Do(func() { close(t.stopCh) })
	<-t.doneCh
}
