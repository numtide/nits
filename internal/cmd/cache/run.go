package cache

import (
	"context"
	"github.com/juju/errors"
	"github.com/numtide/nits/internal/cmd"
	"github.com/numtide/nits/pkg/cache"
)

type runCmd struct{}

func (sc *runCmd) Run() error {
	logger := Cmd.Logging.ToLogger()
	return cmd.Run(logger, func(ctx context.Context) error {
		// process options
		options, err := cacheOptions()
		if err != nil {
			return err
		}

		// create server
		s, err := cache.NewCache(logger, options...)
		if err != nil {
			return errors.Annotate(err, "failed to create server")
		}

		// initialise
		if err := s.Init(); err != nil {
			return errors.Annotate(err, "failed to initialise server")
		}

		// run main loop
		return s.Run(ctx)
	})
}
