package commandtest

import (
	"context"
	"fmt"

	"github.com/i-am-saebyeok/homerepo/internal/command"
)

type Handler func(cmd command.Command) (command.Output, error)

type Fake struct {
	Handlers map[string]Handler
	Fallback command.Runner
	Handled  []command.Command
}

func (f *Fake) Run(ctx context.Context, cmd command.Command) (command.Output, error) {
	handler, ok := f.Handlers[cmd.Name]
	if ok {
		f.Handled = append(f.Handled, cmd)
		return handler(cmd)
	}
	if f.Fallback == nil {
		return command.Output{}, fmt.Errorf("unexpected command: %s", cmd)
	}
	return f.Fallback.Run(ctx, cmd)
}
