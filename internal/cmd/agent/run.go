package agent

import (
	"context"
	"time"

	"github.com/charmbracelet/log"

	"github.com/numtide/nits/internal/cmd"
	"github.com/numtide/nits/pkg/agent"
)

type runCmd struct{}

func (a *runCmd) Run() (err error) {
	level, err := log.ParseLevel(Cmd.LogLevel)
	if err != nil {
		return err
	}

	log.SetLevel(level)
	if err != nil {
		return err
	}
	log.SetFormatter(log.LogfmtFormatter)
	log.SetTimeFormat(time.RFC3339)

	return cmd.Run(func(ctx context.Context) (err error) {
		agent.NatsOptions = &Cmd.Nats
		return agent.Run(ctx)
	})
}
