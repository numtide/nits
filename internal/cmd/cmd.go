package cmd

import (
	"context"
	"os/exec"
	"syscall"
	"time"

	"github.com/ztrue/shutdown"

	"github.com/charmbracelet/log"
)

type (
	Args     = []string
	ArgsList = []Args
)

type LogOptions struct {
	Level string `enum:"debug,info,warn,error,fatal" env:"LOG_LEVEL" default:"warn" help:"Configure logging level."`
}

func (lo *LogOptions) ConfigureLog() {
	log.SetReportTimestamp(true)
	log.SetTimeFormat(time.RFC3339)
	log.SetLevel(log.ParseLevel(lo.Level))
}

func Run(main func(ctx context.Context) error) (err error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	shutdown.Add(cancel)
	go shutdown.Listen(syscall.SIGINT, syscall.SIGTERM)

	return main(ctx)
}

func LogExec(cmd *exec.Cmd) *exec.Cmd {
	log.Debug(cmd.String())
	return cmd
}
