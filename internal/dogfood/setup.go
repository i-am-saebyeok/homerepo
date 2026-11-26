package dogfood

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/i-am-saebyeok/homerepo/internal/command"
	"github.com/i-am-saebyeok/homerepo/internal/herdr"
	"github.com/i-am-saebyeok/homerepo/internal/home"
	"github.com/i-am-saebyeok/homerepo/internal/thread"
)

const (
	modulePath       = "github.com/i-am-saebyeok/homerepo"
	coordinatorLabel = "dogfood-coordinator"
	busyRetryDelay   = 500 * time.Millisecond
)

type runState struct {
	CoordinatorTab string `json:"coordinator_tab"`
}

type session struct {
	app         App
	options     options
	layout      Layout
	workspace   string
	deadline    time.Time
	coordinator herdr.Tab
}

func checkRootIsOurs(layout Layout) error {
	entries, err := os.ReadDir(layout.Root)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("root %s cannot be read: %w", layout.Root, err)
	}
	if len(entries) == 0 || home.IsDir(layout.Marker()) {
		return nil
	}
	return fmt.Errorf("root %s is not empty and was not made by this runner; pass an empty or new folder", layout.Root)
}

func (app App) start(ctx context.Context, chosen options, deadline time.Time) (session, error) {
	s := session{app: app, options: chosen, layout: Layout{Root: chosen.root}, workspace: app.Getenv("HERDR_WORKSPACE_ID"), deadline: deadline}
	steps := []struct {
		name string
		run  func(context.Context) error
	}{
		{"close the previous run's tabs and empty the root", s.reset},
		{"build homerepo from this repository", s.buildHomerepo},
		{"create the empty remote", s.createRemote},
		{"set up the test Home", s.setUpHome},
		{"start the coordinator", s.startCoordinator},
	}
	for _, step := range steps {
		if err := step.run(ctx); err != nil {
			return s, fmt.Errorf("could not %s: %w", step.name, err)
		}
	}
	return s, nil
}

func (s session) reset(ctx context.Context) error {
	tabs, err := s.previousTabs()
	if err != nil {
		return err
	}
	for _, tab := range tabs {
		if err := s.app.herdr().CloseTab(ctx, tab); err != nil {
			return err
		}
	}
	entries, err := os.ReadDir(s.layout.Root)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	for _, entry := range entries {
		if err := os.RemoveAll(filepath.Join(s.layout.Root, entry.Name())); err != nil {
			return err
		}
	}
	return os.MkdirAll(s.layout.Marker(), 0o755)
}

func (s session) previousTabs() ([]string, error) {
	var tabs []string
	content, err := os.ReadFile(s.layout.State())
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}
	if err == nil {
		var state runState
		if err := json.Unmarshal(content, &state); err != nil {
			return nil, fmt.Errorf("%s: %w", s.layout.State(), err)
		}
		tabs = append(tabs, state.CoordinatorTab)
	}
	records, err := thread.Records{HomeRoot: s.layout.Home()}.List()
	if err != nil {
		return nil, err
	}
	for _, record := range records {
		tabs = append(tabs, record.TabID)
	}
	return tabs, nil
}

func (s session) buildHomerepo(ctx context.Context) error {
	_, err := s.app.Runner.Run(ctx, command.Command{Name: "go", Args: []string{"build", "-o", s.layout.Homerepo(), modulePath}})
	return err
}

func (s session) createRemote(ctx context.Context) error {
	_, err := s.git(ctx, s.layout.Root, "init", "--bare", "--quiet", s.layout.Remote(s.options.scenario.Repository))
	return err
}

func (s session) setUpHome(ctx context.Context) error {
	if err := os.MkdirAll(s.layout.Home(), 0o755); err != nil {
		return err
	}
	if _, err := s.git(ctx, s.layout.Home(), "init", "--quiet"); err != nil {
		return err
	}
	if _, err := s.homerepo(ctx, "init"); err != nil {
		return err
	}
	return s.sessionSetup().write()
}

func (s session) sessionSetup() sessionSetup {
	return sessionSetup{layout: s.layout, runtime: s.options.runtime, searchPath: s.app.Getenv("PATH"), runtimeArgs: s.options.runtimeArgs}
}

func (s *session) startCoordinator(ctx context.Context) error {
	tab, err := s.app.herdr().CreateTab(ctx, herdr.TabRequest{Workspace: s.workspace, Cwd: s.layout.Home(), Label: coordinatorLabel})
	if err != nil {
		return err
	}
	s.coordinator = tab
	state, err := json.Marshal(runState{CoordinatorTab: tab.ID})
	if err != nil {
		return err
	}
	if err := os.WriteFile(s.layout.State(), state, 0o644); err != nil {
		return err
	}
	return s.runInPane(ctx, tab.Pane, string(s.options.runtime), s.sessionSetup().launchArgs())
}

func (s session) runInPane(ctx context.Context, pane string, program string, args []string) error {
	for {
		err := s.app.herdr().RunInPane(ctx, pane, program, args)
		if herdr.ErrorCode(err) != "agent_pane_busy" || s.app.remaining(s.deadline) <= 0 {
			return err
		}
		s.app.Sleep(busyRetryDelay)
	}
}

func (s session) git(ctx context.Context, dir string, args ...string) (command.Output, error) {
	return s.app.Runner.Run(ctx, command.Command{Name: "git", Args: args, Dir: dir})
}

func (s session) homerepo(ctx context.Context, args ...string) (command.Output, error) {
	return s.app.Runner.Run(ctx, command.Command{Name: s.layout.Homerepo(), Args: args, Dir: s.layout.Home()})
}

func isFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular()
}
