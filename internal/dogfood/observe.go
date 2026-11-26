package dogfood

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/i-am-saebyeok/homerepo/internal/herdr"
	"github.com/i-am-saebyeok/homerepo/internal/thread"
)

const (
	pollInterval = 2 * time.Second
	startLine    = "Go ahead with the work this thread was created for."
)

var errTimedOut = errors.New("time limit reached")

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
	if errors.Is(err, errTimedOut) || herdr.ErrorCode(err) == "timeout" {
		o.timedOut = true
		o.note("time limit reached: "+format, values...)
		return
	}
	o.note(format+": %v", append(values, err)...)
}

func (s session) observe(ctx context.Context) observation {
	var seen observation
	pane := s.coordinator.Pane
	status, err := s.waitForAgent(ctx, pane)
	if err != nil {
		seen.stop(err, "the coordinator in pane %s did not start", pane)
		return seen
	}
	if status == herdr.AgentBlocked {
		seen.note("the coordinator in pane %s is waiting for a confirmation; start %s once in %s yourself, answer the trust question, and run again", pane, s.options.runtime, s.layout.Home())
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
		seen.stop(err, "no thread was created; see the coordinator in pane %s", pane)
		return seen
	}
	seen.thread = record
	s.startThread(ctx, &seen)
	return seen
}

func (s session) waitForAgent(ctx context.Context, pane string) (herdr.AgentStatus, error) {
	ready := []herdr.AgentStatus{herdr.AgentIdle, herdr.AgentDone, herdr.AgentBlocked}
	for {
		status, err := s.app.herdr().Agent(ctx, pane)
		if err != nil {
			return status, err
		}
		if slices.Contains(ready, status) {
			return status, nil
		}
		if err := s.pause(); err != nil {
			return status, err
		}
	}
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
		if status != herdr.AgentWorking {
			return thread.Record{}, fmt.Errorf("the coordinator stopped as %q", statusLabel(status))
		}
		if err := s.pause(); err != nil {
			return thread.Record{}, err
		}
	}
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
		seen.stop(err, "thread %q in pane %s did not start", seen.thread.Name, pane)
		return
	}
	if status == herdr.AgentBlocked {
		seen.note("thread %q in pane %s is waiting for a confirmation; no keys were sent", seen.thread.Name, pane)
		return
	}
	err = s.app.herdr().Prompt(ctx, herdr.PromptRequest{Pane: pane, Text: startLine, Until: []herdr.AgentStatus{herdr.AgentDone, herdr.AgentIdle, herdr.AgentBlocked}, Timeout: s.promptTimeout()})
	if herdr.ErrorCode(err) == "agent_blocked" {
		seen.note("thread %q in pane %s is waiting for a confirmation; no keys were sent", seen.thread.Name, pane)
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
		seen.note("thread %q in pane %s stopped to wait for a confirmation", seen.thread.Name, pane)
	}
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

func statusLabel(status herdr.AgentStatus) string {
	if status == herdr.AgentAbsent {
		return "no agent"
	}
	return string(status)
}
