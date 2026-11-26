package herdr

import (
	"context"
	"encoding/json"
	"fmt"
)

func (c Client) PaneOfTab(ctx context.Context, tabID string) (string, error) {
	output, err := c.run(ctx, "tab", "get", tabID)
	if ErrorCode(err) == "tab_not_found" {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	var tab struct {
		Result struct {
			Tab struct {
				WorkspaceID string `json:"workspace_id"`
			} `json:"tab"`
		} `json:"result"`
	}
	if err := json.Unmarshal(output.Stdout, &tab); err != nil {
		return "", fmt.Errorf("herdr tab get: unreadable response: %w", err)
	}
	if tab.Result.Tab.WorkspaceID == "" {
		return "", fmt.Errorf("herdr tab get: response has no workspace ID: %s", output.Stdout)
	}
	output, err = c.run(ctx, "pane", "list", "--workspace", tab.Result.Tab.WorkspaceID)
	if err != nil {
		return "", err
	}
	var list struct {
		Result struct {
			Panes []struct {
				PaneID string `json:"pane_id"`
				TabID  string `json:"tab_id"`
			} `json:"panes"`
		} `json:"result"`
	}
	if err := json.Unmarshal(output.Stdout, &list); err != nil {
		return "", fmt.Errorf("herdr pane list: unreadable response: %w", err)
	}
	var panes []string
	for _, pane := range list.Result.Panes {
		if pane.TabID == tabID {
			panes = append(panes, pane.PaneID)
		}
	}
	if len(panes) != 1 {
		return "", nil
	}
	return panes[0], nil
}
