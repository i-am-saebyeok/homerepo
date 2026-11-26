package herdr

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

type AgentStatus string

const (
	AgentIdle    AgentStatus = "idle"
	AgentWorking AgentStatus = "working"
	AgentBlocked AgentStatus = "blocked"
	AgentDone    AgentStatus = "done"
	AgentAbsent  AgentStatus = ""
)

type PromptRequest struct {
	Pane    string
	Text    string
	Until   []AgentStatus
	Timeout time.Duration
}

func (c Client) Agent(ctx context.Context, pane string) (AgentStatus, error) {
	output, err := c.run(ctx, "agent", "get", pane)
	if ErrorCode(err) == "agent_not_found" {
		return AgentAbsent, nil
	}
	if err != nil {
		return AgentAbsent, err
	}
	var response struct {
		Result struct {
			Agent struct {
				Status AgentStatus `json:"agent_status"`
			} `json:"agent"`
		} `json:"result"`
	}
	if err := json.Unmarshal(output.Stdout, &response); err != nil {
		return AgentAbsent, fmt.Errorf("herdr agent get: unreadable response: %w", err)
	}
	return response.Result.Agent.Status, nil
}

func (c Client) Prompt(ctx context.Context, request PromptRequest) error {
	args := []string{"agent", "prompt", request.Pane, request.Text}
	if len(request.Until) > 0 {
		args = append(args, "--wait")
		for _, status := range request.Until {
			args = append(args, "--until", string(status))
		}
		args = append(args, "--timeout", strconv.FormatInt(request.Timeout.Milliseconds(), 10))
	}
	_, err := c.run(ctx, args...)
	return err
}

func (c Client) ReadVisible(ctx context.Context, pane string, lines int) (string, error) {
	output, err := c.run(ctx, "pane", "read", pane, "--source", "visible", "--lines", strconv.Itoa(lines))
	return string(output.Stdout), err
}
