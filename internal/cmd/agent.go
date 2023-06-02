package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/numtide/nits/pkg/agent"
	"github.com/numtide/nits/pkg/util"
)

type agentCmd struct {
	Run  agentRunCmd  `cmd:"" help:"Run an agent." default:"1"`
	Nkey agentNkeyCmd `cmd:"" help:"Get an nkey for a ed25519 key"`
}

type agentRunCmd struct{}

func (a *agentRunCmd) toOptions() ([]agent.Option, error) {
	nats, err := Cli.Nats.toNatsConfig()
	if err != nil {
		return nil, err
	}

	return []agent.Option{
		agent.NatsConfig(nats),
	}, nil
}

func (a *agentRunCmd) Run() error {
	return runCmd(func(ctx context.Context) error {
		// process options
		options, err := Cli.Agent.Run.toOptions()
		if err != nil {
			return err
		}

		// create server
		s, err := agent.NewAgent(options...)
		if err != nil {
			return err
		}

		// initialise
		if err := s.Init(); err != nil {
			return err
		}

		// run main loop
		return s.Run(ctx)
	})
}

type agentNkeyCmd struct {
	KeyFile *os.File `arg:""`
}

func (a *agentNkeyCmd) Run() error {
	signer, err := util.NewSigner(Cli.Agent.Nkey.KeyFile)
	if err != nil {
		return err
	}

	pub, err := util.PublicKeyForSigner(signer)
	if err != nil {
		return err
	}

	fmt.Println(pub)
	return nil
}
