package herdr_test

import (
	"context"
	"errors"
	"testing"

	"github.com/i-am-saebyeok/homerepo/internal/command"
	"github.com/i-am-saebyeok/homerepo/internal/command/commandtest"
	"github.com/i-am-saebyeok/homerepo/internal/herdr"
)

func clientReplying(stdout string, err error) herdr.Client {
	runner := &commandtest.Fake{Handlers: map[string]commandtest.Handler{
		"herdr": func(cmd command.Command) (command.Output, error) {
			return command.Output{Stdout: []byte(stdout)}, err
		},
	}}
	return herdr.Client{Runner: runner}
}

func TestErrorCodeReadsHerdrErrorResponse(t *testing.T) {
	err := commandtest.Fail(command.Command{Name: "herdr"}, `{"error":{"code":"tab_not_found","message":"tab x not found"},"id":"cli:tab:get"}`)
	if code := herdr.ErrorCode(err); code != "tab_not_found" {
		t.Fatalf("code %q", code)
	}
	if code := herdr.ErrorCode(errors.New("plain")); code != "" {
		t.Fatalf("code %q", code)
	}
}

func TestCreateTabRejectsResponseWithoutIDs(t *testing.T) {
	_, err := clientReplying(`{"result":{"tab":{}}}`, nil).CreateTab(context.Background(), herdr.TabRequest{})
	if err == nil {
		t.Fatal("expected an error")
	}
}

func TestCloseTabIgnoresMissingTab(t *testing.T) {
	missing := commandtest.Fail(command.Command{Name: "herdr"}, `{"error":{"code":"tab_not_found","message":"gone"}}`)
	if err := clientReplying("", missing).CloseTab(context.Background(), "w1:t9"); err != nil {
		t.Fatal(err)
	}
}

func TestAgentTreatsMissingAgentAsAbsent(t *testing.T) {
	missing := commandtest.Fail(command.Command{Name: "herdr"}, `{"error":{"code":"agent_not_found","message":"gone"}}`)
	status, err := clientReplying("", missing).Agent(context.Background(), "w1:p2")
	if err != nil || status != herdr.AgentAbsent {
		t.Fatalf("status %q, err %v", status, err)
	}
}

func TestAgentReadsStatus(t *testing.T) {
	status, err := clientReplying(`{"result":{"agent":{"agent_status":"blocked","pane_id":"w1:p2"}}}`, nil).Agent(context.Background(), "w1:p2")
	if err != nil || status != herdr.AgentBlocked {
		t.Fatalf("status %q, err %v", status, err)
	}
}
