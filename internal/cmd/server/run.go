package server

import (
	"context"

	"github.com/juju/errors"
	"github.com/numtide/nits/internal/cmd"
	"github.com/numtide/nits/pkg/server"
)

type runCmd struct{}

func (r *runCmd) Run() error {
	logger := Cmd.Logging.ToLogger()

	return cmd.Run(logger, func(ctx context.Context) error {
		natsConfig, err := Cmd.Nats.ToNatsConfig()
		if err != nil {
			return err
		}

		cacheOptions, err := Cmd.Cache.ToCacheOptions()
		if err != nil {
			return err
		}

		srv, err := server.NewGuvnor(
			logger,
			server.NatsConfig(natsConfig),
			server.CacheOptions(cacheOptions),
		)
		if err != nil {
			return errors.Annotate(err, "failed to create server")
		}

		if err = srv.Init(); err != nil {
			return errors.Annotate(err, "failed to initialise server")
		}

		return srv.Run(ctx)
	})
}
