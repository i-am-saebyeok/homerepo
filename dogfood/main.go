package main

import (
	"context"
	"os"
	"os/exec"
	"time"

	"github.com/i-am-saebyeok/homerepo/internal/command"
	"github.com/i-am-saebyeok/homerepo/internal/dogfood"
)

func main() {
	app := dogfood.App{
		Stdout:   os.Stdout,
		Stderr:   os.Stderr,
		Getenv:   os.Getenv,
		LookPath: exec.LookPath,
		Runner:   command.Exec{},
		Now:      time.Now,
		Sleep:    time.Sleep,
	}
	os.Exit(app.Run(context.Background(), os.Args[1:]))
}
