package deploy

import (
	"context"

	"github.com/charmbracelet/log"
	"github.com/numtide/nits/pkg/types"
)

type Deployer int

const (
	DeployerNoOp Deployer = iota
	DeployerNixos
)

func (a Deployer) String() string {
	switch a {
	case DeployerNoOp:
		return "noop"
	case DeployerNixos:
		return "nixos"
	}
	return "unknown"
}

func ParseDeployer(str string) Deployer {
	switch str {
	case "noop":
		return DeployerNoOp
	case "nixos":
		return DeployerNixos
	default:
		return DeployerNoOp
	}
}

type HandlerFunc func(*types.Deployment, context.Context) error

func (f HandlerFunc) Apply(config *types.Deployment, ctx context.Context) error {
	return f(config, ctx)
}

type Handler interface {
	Apply(config *types.Deployment, ctx context.Context) error
}

func NoOpHandler(_ *types.Deployment, ctx context.Context) error {
	log.FromContext(ctx).Info("NoOp handler, doing nothing")
	return nil
}
