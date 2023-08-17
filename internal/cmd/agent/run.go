package agent

import (
	"context"

	"github.com/numtide/nits/internal/cmd"
	"github.com/numtide/nits/pkg/agent"
	"github.com/numtide/nits/pkg/agent/deploy"
)

type runCmd struct {
	CacheProxy          cmd.CacheProxyOptions `embed:"" prefix:"cache-proxy-"`
	Deployer            string                `enum:"noop,nixos" env:"NITS_AGENT_DEPLOYER" default:"nixos" help:"Configure deployer to use."`
	SubjectPrefixFormat string                `name:"subject-prefix-format" env:"NITS_AGENT_SUBJECT_PREFIX_FORMAT" default:"NITS.AGENT.%s"`
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

	cacheProxyConfig, err := a.CacheProxy.ToCacheProxyConfig()
	if err != nil {
		return err
	}

	return cmd.Run(logger, func(ctx context.Context) error {
		a := agent.Agent{
			NatsConfig:          natsConfig,
			Deployer:            deploy.ParseDeployer(a.Deployer),
			SubjectPrefixFormat: a.SubjectPrefixFormat,
			CacheProxyConfig:    cacheProxyConfig,
		}
		return a.Run(ctx, logger)
	})
}
