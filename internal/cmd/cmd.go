package cmd

import (
	"context"
	"github.com/ztrue/shutdown"
	"os/exec"
	"syscall"
	"time"

	"github.com/charmbracelet/log"
)

type (
	Args     = []string
	ArgsList = []Args
)

type LogOptions struct {
	Verbosity    int  `name:"verbose" short:"v" type:"counter" default:"0" env:"LOG_LEVEL" help:"Set the verbosity of logs e.g. -vv."`
	LogTimestamp bool `default:"false" env:"LOG_TIMESTAMP" help:"Add timestamp to log output"`
}

func (lo *LogOptions) ConfigureLog() error {
	log.SetReportTimestamp(lo.LogTimestamp)
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
