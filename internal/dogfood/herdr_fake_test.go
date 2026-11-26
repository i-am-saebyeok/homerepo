package dogfood_test

import (
	"fmt"
	"slices"
	"strings"

	"github.com/i-am-saebyeok/homerepo/internal/command"
	"github.com/i-am-saebyeok/homerepo/internal/command/commandtest"
	"github.com/i-am-saebyeok/homerepo/internal/herdr"
	"github.com/i-am-saebyeok/homerepo/internal/herdr/herdrtest"
)

type prompt struct {
	pane string
	text string
	args []string
}

type fakeHerdr struct {
	world           *world
	created         int
	openTabs        map[string]string
	closedTabs      []string
	statuses        map[string][]herdr.AgentStatus
	coordinatorPane string
	threadPane      string
	busyRuns        int
	paneRuns        [][]string
	prompts         []prompt
	keysSent        bool
	handleTabCreate func(pane string)
	screens         map[string]string
}

func newFakeHerdr(w *world) *fakeHerdr {
	return &fakeHerdr{world: w, openTabs: map[string]string{}, statuses: map[string][]herdr.AgentStatus{}, screens: map[string]string{}}
}

func fail(cmd command.Command, code string) error {
	return commandtest.Fail(cmd, `{"error":{"code":"`+code+`","message":"fake"}}`)
}

func reply(json string) (command.Output, error) {
	return command.Output{Stdout: []byte(json)}, nil
}

func (h *fakeHerdr) newTab() herdr.Tab {
	h.created++
	tab := herdr.Tab{ID: fmt.Sprintf("w1:t%d", h.created), Pane: fmt.Sprintf("w1:p%d", h.created)}
	h.openTabs[tab.ID] = tab.Pane
	return tab
}

func (h *fakeHerdr) status(pane string) herdr.AgentStatus {
	queue := h.statuses[pane]
	if len(queue) == 0 {
		return herdr.AgentAbsent
	}
	if len(queue) > 1 {
		h.statuses[pane] = queue[1:]
	}
	return queue[0]
}

func (h *fakeHerdr) handle(cmd command.Command) (command.Output, error) {
	switch strings.Join(cmd.Args[:2], " ") {
	case "tab create":
		tab := h.newTab()
		h.statuses[tab.Pane] = []herdr.AgentStatus{herdr.AgentAbsent, herdr.AgentIdle}
		if slices.Contains(cmd.Args, "dogfood-coordinator") {
			h.coordinatorPane = tab.Pane
		} else {
			h.threadPane = tab.Pane
		}
		if h.handleTabCreate != nil {
			h.handleTabCreate(tab.Pane)
		}
		return reply(`{"result":{"tab":{"tab_id":"` + tab.ID + `"},"root_pane":{"pane_id":"` + tab.Pane + `"}}}`)
	case "tab get":
		if _, ok := h.openTabs[cmd.Args[2]]; !ok {
			return command.Output{}, fail(cmd, "tab_not_found")
		}
		return reply(`{"result":{"tab":{"tab_id":"` + cmd.Args[2] + `","workspace_id":"` + workspaceID + `"}}}`)
	case "tab close":
		h.closedTabs = append(h.closedTabs, cmd.Args[2])
		if _, ok := h.openTabs[cmd.Args[2]]; !ok {
			return command.Output{}, fail(cmd, "tab_not_found")
		}
		delete(h.openTabs, cmd.Args[2])
		return reply(`{"result":{}}`)
	case "pane run":
		if h.busyRuns > 0 {
			h.busyRuns--
			return command.Output{}, fail(cmd, "pane_send_failed")
		}
		h.paneRuns = append(h.paneRuns, append([]string{cmd.Args[2]}, herdrtest.ShellWords(h.world.t, cmd.Args[3:])...))
		return reply(`{"result":{}}`)
	case "pane list":
		var panes []string
		for tab, pane := range h.openTabs {
			panes = append(panes, `{"pane_id":"`+pane+`","tab_id":"`+tab+`"}`)
		}
		return reply(`{"result":{"panes":[` + strings.Join(panes, ",") + `]}}`)
	case "agent get":
		status := h.status(cmd.Args[2])
		if status == herdr.AgentAbsent {
			return command.Output{}, fail(cmd, "agent_not_found")
		}
		return reply(`{"result":{"agent":{"agent_status":"` + string(status) + `"}}}`)
	case "pane read":
		return reply(h.screens[cmd.Args[2]])
	case "agent prompt":
		return h.prompt(cmd)
	case "agent send-keys":
		h.keysSent = true
		return reply(`{"result":{}}`)
	}
	return command.Output{}, fmt.Errorf("unexpected herdr call: %v", cmd.Args)
}

func (h *fakeHerdr) prompt(cmd command.Command) (command.Output, error) {
	pane := cmd.Args[2]
	h.prompts = append(h.prompts, prompt{pane: pane, text: cmd.Args[3], args: cmd.Args[4:]})
	if h.status(pane) == herdr.AgentBlocked {
		return command.Output{}, fail(cmd, "agent_blocked")
	}
	switch pane {
	case h.coordinatorPane:
		h.statuses[pane] = []herdr.AgentStatus{herdr.AgentWorking, herdr.AgentDone}
		h.world.coordinate(h.world)
	case h.threadPane:
		h.statuses[pane] = []herdr.AgentStatus{herdr.AgentDone}
		if err := h.world.work(h.world); err != nil {
			return command.Output{}, fail(cmd, err.Error())
		}
	}
	return reply(`{"result":{}}`)
}

func (h *fakeHerdr) promptsTo(pane string) []prompt {
	var found []prompt
	for _, sent := range h.prompts {
		if sent.pane == pane {
			found = append(found, sent)
		}
	}
	return found
}
