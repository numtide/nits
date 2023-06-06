package guvnor

import (
	"context"
	"github.com/juju/errors"
	"github.com/numtide/nits/internal/cmd"
	"github.com/numtide/nits/pkg/guvnor"
)

type runCmd struct {
}

func (r *runCmd) Run() error {
	logger := Cmd.Logging.ToLogger()

	return cmd.Run(logger, func(ctx context.Context) error {

		natsConfig, err := Cmd.Nats.ToNatsConfig()
		if err != nil {
			return err
		}

		srv, err := guvnor.NewGuvnor(
			logger,
			guvnor.NatsConfig(natsConfig),
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
