package herdr_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

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

func clientRejectingRuns(rejections int, code string) (herdr.Client, *[]time.Duration, *int) {
	var slept []time.Duration
	runs := 0
	runner := &commandtest.Fake{Handlers: map[string]commandtest.Handler{
		"herdr": func(cmd command.Command) (command.Output, error) {
			runs++
			if runs <= rejections {
				return command.Output{}, commandtest.Fail(cmd, `{"error":{"code":"`+code+`","message":"wait for the shell prompt and retry"}}`)
			}
			return command.Output{Stdout: []byte(`{"result":{}}`)}, nil
		},
	}}
	return herdr.Client{Runner: runner, Sleep: func(d time.Duration) { slept = append(slept, d) }}, &slept, &runs
}

func TestRunInPaneSendsAgainUntilAccepted(t *testing.T) {
	for _, code := range []string{"agent_pane_busy", "pane_send_failed"} {
		client, slept, runs := clientRejectingRuns(4, code)
		if err := client.RunInPane(context.Background(), "w1:p1", "codex", []string{"--add-dir", "/a b"}); err != nil {
			t.Fatalf("%s: %v", code, err)
		}
		if *runs != 5 || len(*slept) != 4 || (*slept)[0] != herdr.SendRetryInterval {
			t.Fatalf("%s: runs %d, slept %v", code, *runs, *slept)
		}
	}
}

func TestRunInPaneGivesUpAfterLimit(t *testing.T) {
	client, slept, _ := clientRejectingRuns(1000, "pane_send_failed")
	err := client.RunInPane(context.Background(), "w1:p1", "codex", nil)
	if herdr.ErrorCode(err) != "pane_send_failed" {
		t.Fatalf("err %v", err)
	}
	var total time.Duration
	for _, d := range *slept {
		total += d
	}
	if total != herdr.SendRetryLimit {
		t.Fatalf("waited %s", total)
	}
}

func TestRunInPaneReturnsOtherErrorsAtOnce(t *testing.T) {
	client, slept, runs := clientRejectingRuns(1000, "pane_not_found")
	if err := client.RunInPane(context.Background(), "w1:p1", "codex", nil); herdr.ErrorCode(err) != "pane_not_found" || *runs != 1 || len(*slept) != 0 {
		t.Fatalf("err %v, runs %d, slept %v", err, *runs, *slept)
	}
}

func TestReadVisibleAsksForTheVisibleScreen(t *testing.T) {
	var asked []string
	runner := &commandtest.Fake{Handlers: map[string]commandtest.Handler{
		"herdr": func(cmd command.Command) (command.Output, error) {
			asked = cmd.Args
			return command.Output{Stdout: []byte("Allow this command?\n")}, nil
		},
	}}
	screen, err := herdr.Client{Runner: runner}.ReadVisible(context.Background(), "w1:p2", 12)
	if err != nil || screen != "Allow this command?\n" {
		t.Fatalf("screen %q, err %v", screen, err)
	}
	if strings.Join(asked, " ") != "pane read w1:p2 --source visible --lines 12" {
		t.Fatalf("asked %v", asked)
	}
}
