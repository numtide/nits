package agent

import (
	"context"

	"github.com/numtide/nits/internal/cmd"
	"github.com/numtide/nits/pkg/agent"
	"github.com/numtide/nits/pkg/agent/deploy"
)

type runCmd struct {
	CacheProxy cmd.CacheProxyOptions `embed:"" prefix:"cache-proxy-"`
	Deployer   string                `enum:"noop,nixos" env:"NITS_AGENT_DEPLOYER" default:"nixos" help:"Configure deployer to use."`
}

func (a *runCmd) Run() (err error) {
	logger, err := Cmd.Logging.ToLogger()
	if err != nil {
		return err
	}

	cacheProxyConfig, err := a.CacheProxy.ToCacheProxyConfig()
	if err != nil {
		return err
	}

	return cmd.Run(logger, func(ctx context.Context) error {
		a := agent.Agent{
			NatsOptions:      &Cmd.Nats,
			Deployer:         deploy.ParseDeployer(a.Deployer),
			CacheProxyConfig: cacheProxyConfig,
		}
		return a.Run(ctx, logger)
	})
}
