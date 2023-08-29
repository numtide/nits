package service

import (
	"context"

	"github.com/numtide/nits/pkg/agent/service/info"
)

func Init(ctx context.Context) (err error) {
	return info.Init(ctx)
}
