package service

import (
	"context"
	"github.com/numtide/nits/pkg/agent/service/nixos"

	"github.com/numtide/nits/pkg/agent/service/cmd"

	"github.com/numtide/nits/pkg/agent/service/info"
)

func Init(ctx context.Context) (err error) {
	if err = info.Init(ctx); err != nil {
		return
	} else if err = cmd.Init(ctx); err != nil {
		return
	}

	return nixos.Init(ctx)
}
