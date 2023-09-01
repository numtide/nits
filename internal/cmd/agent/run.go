package agent

import (
	"context"

	"github.com/charmbracelet/log"

	"github.com/numtide/nits/internal/cmd"
	"github.com/numtide/nits/pkg/agent"
)

type runCmd struct{}

func (a *runCmd) Run() (err error) {
	log.SetLevel(log.ParseLevel(Cmd.LogLevel))
	log.SetFormatter(log.LogfmtFormatter)
	return cmd.Run(func(ctx context.Context) (err error) {
		agent.NatsOptions = &Cmd.Nats
		return agent.Run(ctx)
	})
}
