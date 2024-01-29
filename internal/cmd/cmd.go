package cmd

import (
	"context"
	"fmt"
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

func (lo *LogOptions) ConfigureLog() error {
	log.SetReportTimestamp(true)
	log.SetTimeFormat(time.RFC3339)

	level, err := log.ParseLevel(lo.Level)
	if err != nil {
		return fmt.Errorf("%w: failed to parse log level %v", err, lo.Level)
	}

	log.SetLevel(level)
	return nil
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
