package cache

import (
	"context"
	"time"

	"github.com/juju/errors"
	"github.com/numtide/nits/internal/cmd"
	"github.com/numtide/nits/pkg/cache"
)

type gcCmd struct {
	TimeWindow time.Duration `env:"NITS_CACHE_GC_TIME_WINDOW" default:"15120h"` // 90 days
}

func (sc *gcCmd) Run() error {
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

		cutOff := time.Now().Add(sc.TimeWindow * -1)

		// run gc
		return s.GarbageCollect(ctx, cutOff)
	})
}
