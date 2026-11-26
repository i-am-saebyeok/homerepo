package main

import (
	"context"
	"os"
	"time"

	"github.com/i-am-saebyeok/homerepo/internal/cli"
	"github.com/i-am-saebyeok/homerepo/internal/command"
)

func main() {
	app := cli.App{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Getenv: os.Getenv,
		Getwd:  os.Getwd,
		Runner: command.Exec{},
		Sleep:  time.Sleep,
		Now:    time.Now,
	}
	os.Exit(app.Run(context.Background(), os.Args[1:]))
}
