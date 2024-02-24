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
	Verbosity int `name:"verbose" short:"v" type:"counter" default:"0" env:"LOG_LEVEL" help:"Set the verbosity of logs e.g. -vv."`
}

func (lo *LogOptions) ConfigureLog() error {
	log.SetReportTimestamp(true)
	log.SetTimeFormat(time.RFC3339)

	if lo.Verbosity == 0 {
		log.SetLevel(log.WarnLevel)
	} else if lo.Verbosity == 1 {
		log.SetLevel(log.InfoLevel)
	} else if lo.Verbosity > 1 {
		log.SetLevel(log.DebugLevel)
	}

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
