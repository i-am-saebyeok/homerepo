package dogfood

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/i-am-saebyeok/homerepo/internal/herdr"
	"github.com/i-am-saebyeok/homerepo/internal/thread"
)

const (
	pollInterval = 2 * time.Second
	startupLimit = time.Minute
	screenLines  = 15
	startLine    = "Go ahead with the work this thread was created for."
)

var (
	errTimedOut     = errors.New("time limit reached")
	errExited       = errors.New("the runtime exited and the pane is back at its shell")
	errNeverStarted = fmt.Errorf("no runtime appeared in the pane within %s", startupLimit)
	errWaiting      = errors.New("waiting for a confirmation")
)

type observation struct {
	thread        thread.Record
	threadPane    string
	threadStarted bool
	timedOut      bool
	problems      []string
}

func (o *observation) note(format string, values ...any) {
	o.problems = append(o.problems, fmt.Sprintf(format, values...))
}

func (o *observation) stop(err error, format string, values ...any) {
	o.stopShowing(err, "", format, values...)
}

func (o *observation) stopShowing(err error, screen string, format string, values ...any) {
	if errors.Is(err, errTimedOut) || herdr.ErrorCode(err) == "timeout" {
		o.timedOut = true
		o.note("time limit reached: "+format, values...)
		return
	}
	o.note(format+": %v%s", append(values, err, screen)...)
}

func (s session) observe(ctx context.Context) observation {
	var seen observation
	pane := s.coordinator.Pane
	status, err := s.waitForAgent(ctx, pane)
	if err != nil {
		seen.stopShowing(err, s.screen(ctx, pane, err), "the coordinator in pane %s did not start", pane)
		return seen
	}
	if status == herdr.AgentBlocked {
		seen.note("the coordinator in pane %s is waiting for a confirmation before it got the scenario; if it is asking whether to trust %s, start %s there once yourself, answer the trust question, and run again%s", pane, s.layout.Home(), s.options.runtime, s.screen(ctx, pane, errWaiting))
		return seen
	}
	instructions, err := s.options.scenario.Instructions(s.layout)
	if err != nil {
		seen.stop(err, "the scenario could not be read")
		return seen
	}
	if err := s.app.herdr().Prompt(ctx, herdr.PromptRequest{Pane: pane, Text: instructions, Until: []herdr.AgentStatus{herdr.AgentWorking, herdr.AgentBlocked}, Timeout: s.promptTimeout()}); err != nil {
		seen.stop(err, "the coordinator in pane %s did not take the scenario", pane)
		return seen
	}
	record, err := s.waitForThread(ctx)
	if err != nil {
		seen.stopShowing(err, s.screen(ctx, pane, err), "no thread was created; see the coordinator in pane %s", pane)
		return seen
	}
	seen.thread = record
	s.startThread(ctx, &seen)
	return seen
}

func (s session) waitForAgent(ctx context.Context, pane string) (herdr.AgentStatus, error) {
	ready := []herdr.AgentStatus{herdr.AgentIdle, herdr.AgentDone, herdr.AgentBlocked}
	launched := s.app.Now()
	appeared := false
	for {
		status, err := s.app.herdr().Agent(ctx, pane)
		if err != nil {
			return status, err
		}
		if slices.Contains(ready, status) {
			return status, nil
		}
		if err := startupFailure(status, appeared, s.app.Now().Sub(launched)); err != nil {
			return status, err
		}
		appeared = appeared || status != herdr.AgentAbsent
		if err := s.pause(); err != nil {
			return status, err
		}
	}
}

func startupFailure(status herdr.AgentStatus, appeared bool, waited time.Duration) error {
	if status != herdr.AgentAbsent {
		return nil
	}
	if appeared {
		return errExited
	}
	if waited >= startupLimit {
		return errNeverStarted
	}
	return nil
}

func (s session) waitForThread(ctx context.Context) (thread.Record, error) {
	for {
		status, err := s.app.herdr().Agent(ctx, s.coordinator.Pane)
		if err != nil {
			return thread.Record{}, err
		}
		records, err := thread.Records{HomeRoot: s.layout.Home()}.List()
		if err != nil {
			return thread.Record{}, err
		}
		if len(records) > 0 {
			return records[0], nil
		}
		if err := stoppedWorking(status); err != nil {
			return thread.Record{}, err
		}
		if err := s.pause(); err != nil {
			return thread.Record{}, err
		}
	}
}

func stoppedWorking(status herdr.AgentStatus) error {
	if status == herdr.AgentWorking {
		return nil
	}
	if status == herdr.AgentAbsent {
		return errExited
	}
	if status == herdr.AgentBlocked {
		return errWaiting
	}
	return fmt.Errorf("the coordinator stopped as %q", status)
}

func (s session) startThread(ctx context.Context, seen *observation) {
	pane, err := s.waitForPane(ctx, seen.thread.TabID)
	if err != nil {
		seen.stop(err, "thread %q has no pane in tab %s", seen.thread.Name, seen.thread.TabID)
		return
	}
	seen.threadPane = pane
	status, err := s.waitForAgent(ctx, pane)
	if err != nil {
		seen.stopShowing(err, s.screen(ctx, pane, err), "thread %q in pane %s did not start", seen.thread.Name, pane)
		return
	}
	if status == herdr.AgentBlocked {
		seen.note("thread %q in pane %s is waiting for a confirmation; no keys were sent%s", seen.thread.Name, pane, s.screen(ctx, pane, errWaiting))
		return
	}
	err = s.app.herdr().Prompt(ctx, herdr.PromptRequest{Pane: pane, Text: startLine, Until: []herdr.AgentStatus{herdr.AgentDone, herdr.AgentIdle, herdr.AgentBlocked}, Timeout: s.promptTimeout()})
	if herdr.ErrorCode(err) == "agent_blocked" {
		seen.note("thread %q in pane %s is waiting for a confirmation; no keys were sent%s", seen.thread.Name, pane, s.screen(ctx, pane, errWaiting))
		return
	}
	seen.threadStarted = true
	if err != nil {
		seen.stop(err, "thread %q in pane %s did not finish", seen.thread.Name, pane)
		return
	}
	status, err = s.app.herdr().Agent(ctx, pane)
	if err != nil {
		seen.stop(err, "the state of thread %q could not be read", seen.thread.Name)
		return
	}
	if status == herdr.AgentBlocked {
		seen.note("thread %q in pane %s stopped to wait for a confirmation%s", seen.thread.Name, pane, s.screen(ctx, pane, errWaiting))
	}
}

func (s session) screen(ctx context.Context, pane string, cause error) string {
	if !errors.Is(cause, errWaiting) && !errors.Is(cause, errExited) && !errors.Is(cause, errNeverStarted) {
		return ""
	}
	text, err := s.app.herdr().ReadVisible(ctx, pane, screenLines)
	if err != nil {
		return fmt.Sprintf("\n  (the screen could not be read: %v)", err)
	}
	lines := strings.Split(strings.TrimRight(text, " \n"), "\n")
	if len(lines) > screenLines {
		lines = lines[len(lines)-screenLines:]
	}
	return "\n  last lines of the screen:\n  | " + strings.Join(lines, "\n  | ")
}

func (s session) waitForPane(ctx context.Context, tabID string) (string, error) {
	for {
		pane, err := s.app.herdr().PaneOfTab(ctx, tabID)
		if err != nil || pane != "" {
			return pane, err
		}
		if err := s.pause(); err != nil {
			return "", err
		}
	}
}

func (s session) promptTimeout() time.Duration {
	return max(s.app.remaining(s.deadline), time.Millisecond)
}

func (s session) pause() error {
	if s.app.remaining(s.deadline) <= 0 {
		return errTimedOut
	}
	s.app.Sleep(min(pollInterval, s.app.remaining(s.deadline)))
	return nil
}
