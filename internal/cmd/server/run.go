package server

import (
	"context"

	"github.com/numtide/nits/internal/cmd"
	"github.com/numtide/nits/pkg/server"
)

type runCmd struct {
	Cache        cmd.CacheOptions `embed:"" prefix:"cache-"`
	CacheAddress string           `env:"NITS_CACHE_ADDRESS" default:"localhost:3000"`
}

func (r *runCmd) Run() error {
	logger, err := Cmd.Logging.ToLogger()
	if err != nil {
		return err
	}

	// todo
	//if Cmd.Nats.InboxFormat == "" {
	//	Cmd.Nats.InboxFormat = "nits.server.%s.inbox"
	//}

	return cmd.Run(logger, func(ctx context.Context) error {

		cacheOptions, err := r.Cache.ToCacheOptions()
		if err != nil {
			return err
		}

		srv := server.Server{
			DataDir:      Cmd.DataDir,
			CacheOptions: cacheOptions,
			CacheAddress: r.CacheAddress,
		}

		return srv.Run(ctx, logger)
	})
}
