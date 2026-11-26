package herdr

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/i-am-saebyeok/homerepo/internal/command"
)

type Client struct {
	Runner command.Runner
}

type Tab struct {
	ID   string
	Pane string
}

type TabRequest struct {
	Workspace   string
	Cwd         string
	Label       string
	Environment []string
}

func (c Client) CreateTab(ctx context.Context, request TabRequest) (Tab, error) {
	args := []string{"tab", "create", "--workspace", request.Workspace, "--cwd", request.Cwd, "--label", request.Label}
	for _, variable := range request.Environment {
		args = append(args, "--env", variable)
	}
	output, err := c.run(ctx, append(args, "--no-focus")...)
	if err != nil {
		return Tab{}, err
	}
	var response struct {
		Result struct {
			Tab struct {
				TabID string `json:"tab_id"`
			} `json:"tab"`
			RootPane struct {
				PaneID string `json:"pane_id"`
			} `json:"root_pane"`
		} `json:"result"`
	}
	if err := json.Unmarshal(output.Stdout, &response); err != nil {
		return Tab{}, fmt.Errorf("herdr tab create: unreadable response: %w", err)
	}
	if response.Result.Tab.TabID == "" || response.Result.RootPane.PaneID == "" {
		return Tab{}, fmt.Errorf("herdr tab create: response has no tab or pane ID: %s", output.Stdout)
	}
	return Tab{ID: response.Result.Tab.TabID, Pane: response.Result.RootPane.PaneID}, nil
}

func (c Client) RunInPane(ctx context.Context, pane string, program string, args []string) error {
	_, err := c.run(ctx, append([]string{"pane", "run", pane, program}, args...)...)
	return err
}

func (c Client) TabExists(ctx context.Context, tabID string) (bool, error) {
	_, err := c.run(ctx, "tab", "get", tabID)
	if ErrorCode(err) == "tab_not_found" {
		return false, nil
	}
	return err == nil, err
}

func (c Client) CloseTab(ctx context.Context, tabID string) error {
	_, err := c.run(ctx, "tab", "close", tabID)
	if ErrorCode(err) == "tab_not_found" {
		return nil
	}
	return err
}

func ErrorCode(err error) string {
	var failure *command.Failure
	if !errors.As(err, &failure) {
		return ""
	}
	var response struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if json.Unmarshal([]byte(failure.Stderr), &response) != nil {
		return ""
	}
	return response.Error.Code
}

func (c Client) run(ctx context.Context, args ...string) (command.Output, error) {
	return c.Runner.Run(ctx, command.Command{Name: "herdr", Args: args})
}
