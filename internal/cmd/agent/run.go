package agent

import (
	"context"

	"github.com/numtide/nits/internal/cmd"
	"github.com/numtide/nits/pkg/agent"
)

type runCmd struct {
	DryRun           bool   `name:"dry-run" env:"NITS_AGENT_DRY_RUN" default:"false"`
	LogSubjectFormat string `name:"log-subject-format" env:"NITS_AGENT_LOG_SUBJECT_FORMAT"`
}

func (a *runCmd) toOptions() ([]agent.Option, error) {
	nats, err := Cmd.Nats.ToNatsConfig()
	if err != nil {
		return nil, err
	}

	opts := []agent.Option{
		agent.NatsConfig(nats),
		agent.SwitchDryRun(Cmd.Run.DryRun),
	}
	if Cmd.Run.LogSubjectFormat != "" {
		opts = append(opts, agent.LogSubjectFormat(Cmd.Run.LogSubjectFormat))
	}

	return opts, nil
}

func (a *runCmd) Run() error {
	logger := Cmd.Logging.ToLogger()
	return cmd.Run(logger, func(ctx context.Context) error {
		// process options
		options, err := Cmd.Run.toOptions()
		if err != nil {
			return err
		}

		// create server
		s, err := agent.NewAgent(logger, options...)
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
