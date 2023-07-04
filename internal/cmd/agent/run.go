package agent

import (
	"context"

	"github.com/numtide/nits/internal/cmd"
	"github.com/numtide/nits/pkg/agent"
)

type runCmd struct {
	DryRun              bool   `name:"dry-run" env:"NITS_AGENT_DRY_RUN" default:"false"`
	SubjectPrefixFormat string `name:"subject-prefix-format" env:"NITS_AGENT_SUBJECT_PREFIX_FORMAT" default:"nits.agent.%s"`
}

func (a *runCmd) Run() error {
	logger, err := Cmd.Logging.ToLogger()
	if err != nil {
		return err
	}

	natsConfig, err := Cmd.Nats.ToNatsConfig()
	if err != nil {
		return err
	}

	return cmd.Run(logger, func(ctx context.Context) error {
		a := agent.Agent{
			DryRun:              a.DryRun,
			NatsConfig:          natsConfig,
			SubjectPrefixFormat: a.SubjectPrefixFormat,
		}
		return a.Run(ctx, logger)
	})
}
