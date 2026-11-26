package herdr_test

import (
	"context"
	"testing"

	"github.com/i-am-saebyeok/homerepo/internal/command"
	"github.com/i-am-saebyeok/homerepo/internal/command/commandtest"
	"github.com/i-am-saebyeok/homerepo/internal/herdr"
)

func clientWithPanes(tabResponse string, tabErr error, panes string) (herdr.Client, *commandtest.Fake) {
	runner := &commandtest.Fake{Handlers: map[string]commandtest.Handler{
		"herdr": func(cmd command.Command) (command.Output, error) {
			if cmd.Args[0] == "tab" {
				return command.Output{Stdout: []byte(tabResponse)}, tabErr
			}
			return command.Output{Stdout: []byte(`{"result":{"panes":[` + panes + `]}}`)}, nil
		},
	}}
	return herdr.Client{Runner: runner}, runner
}

func TestPaneOfTabFindsTheTabsOnlyPane(t *testing.T) {
	client, runner := clientWithPanes(`{"result":{"tab":{"tab_id":"w2:t4","workspace_id":"w2"}}}`, nil,
		`{"pane_id":"w2:p1","tab_id":"w2:t1"},{"pane_id":"w2:p7","tab_id":"w2:t4"}`)
	pane, err := client.PaneOfTab(context.Background(), "w2:t4")
	if err != nil || pane != "w2:p7" {
		t.Fatalf("pane %q, err %v", pane, err)
	}
	if got := runner.Handled[1].Args; len(got) != 4 || got[3] != "w2" {
		t.Fatalf("pane list call %q", got)
	}
}

func TestPaneOfTabIsEmptyForMissingTabOrSeveralPanes(t *testing.T) {
	missing := commandtest.Fail(command.Command{Name: "herdr"}, `{"error":{"code":"tab_not_found","message":"gone"}}`)
	client, _ := clientWithPanes("", missing, "")
	if pane, err := client.PaneOfTab(context.Background(), "w2:t4"); err != nil || pane != "" {
		t.Fatalf("missing tab: pane %q, err %v", pane, err)
	}
	client, _ = clientWithPanes(`{"result":{"tab":{"workspace_id":"w2"}}}`, nil, `{"pane_id":"w2:p7","tab_id":"w2:t4"},{"pane_id":"w2:p8","tab_id":"w2:t4"}`)
	if pane, err := client.PaneOfTab(context.Background(), "w2:t4"); err != nil || pane != "" {
		t.Fatalf("split tab: pane %q, err %v", pane, err)
	}
}

func TestPaneOfTabFailsWithoutWorkspace(t *testing.T) {
	client, _ := clientWithPanes(`{"result":{"tab":{}}}`, nil, "")
	if _, err := client.PaneOfTab(context.Background(), "w2:t4"); err == nil {
		t.Fatal("expected an error")
	}
}
