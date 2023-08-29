package agent

import (
	"context"

	"github.com/charmbracelet/log"

	"github.com/numtide/nits/internal/cmd"
	"github.com/numtide/nits/pkg/agent"
	"github.com/numtide/nits/pkg/agent/deploy"
)

type runCmd struct {
	Deployer string `enum:"noop,nixos" env:"NITS_AGENT_DEPLOYER" default:"nixos" help:"Configure deployer to use."`
}

func (a *runCmd) Run() (err error) {
	var logger *log.Logger
	if logger, err = Cmd.Logging.ToLogger(); err != nil {
		return
	}

	return cmd.Run(logger, func(ctx context.Context) error {
		agent.Log = logger
		agent.Deployer = deploy.ParseDeployer(a.Deployer)
		agent.NatsOptions = &Cmd.Nats
		return agent.Run(ctx)
	})
}
