package main

import (
	"context"
	"os"

	"github.com/i-am-saebyeok/homerepo/internal/cli"
	"github.com/i-am-saebyeok/homerepo/internal/command"
)

func main() {
	app := cli.App{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Getwd:  os.Getwd,
		Runner: command.Exec{},
	}
	os.Exit(app.Run(context.Background(), os.Args[1:]))
}
