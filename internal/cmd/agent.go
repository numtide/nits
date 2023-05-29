package cmd

import (
	"context"
	"github.com/juju/errors"
	"github.com/numtide/nits/pkg/agent"
	"os"
)

type agentCmd struct {
	HostKeyFile *os.File `name:"" env:"NITS_AGENT_HOST_KEY_FILE" default:"./key"`
}

func (a *agentCmd) toOptions() ([]agent.Option, error) {

	nats, err := natsConfig()
	if err != nil {
		return nil, err
	}

	return []agent.Option{
		agent.NatsConfig(nats),
	}, nil
}

func (a *agentCmd) Run() error {
	return runCmd(func(ctx context.Context) error {
		// process options
		options, err := Cli.Agent.toOptions()
		if err != nil {
			return err
		}

		// create server
		s, err := agent.NewAgent(logger, options...)
		if err != nil {
			return errors.Annotate(err, "failed to create server")
		}

		// initialise
		if err := s.Init(); err != nil {
			return errors.Annotate(err, "failed to initialise server")
		}

		// run main loop
		return s.Run(ctx)
	})
}
