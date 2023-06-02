package cmd

import (
	"context"
	log "github.com/inconshreveable/log15"
	"os"
	"time"

	"github.com/juju/errors"
	"github.com/numtide/nits/pkg/cache"
)

type cacheCmd struct {
	StoreDir       string   `env:"NITS_CACHE_STORE_DIR" default:"/nix/store"`
	WantMassQuery  bool     `env:"NITS_CACHE_WANT_MASS_QUERY" default:"true"`
	Priority       int      `env:"NITS_CACHE_PRIORITY" default:"1"`
	PrivateKeyFile *os.File `env:"NITS_CACHE_PRIVATE_KEY_FILE"`

	Run cacheRunCmd `cmd:"" help:"Run a binary cache."`
	GC  cacheGcCmd  `cmd:"" help:"Garbage collect the binary cache."`
}

type cacheRunCmd struct{}

type cacheGcCmd struct {
	TimeWindow time.Duration `env:"NITS_CACHE_GC_TIME_WINDOW" default:"15120h"` // 90 days
}

func (sc *cacheCmd) toOptions() ([]cache.Option, error) {
	c := Cli.Cache

	nats, err := Cli.Nats.toNatsConfig()
	if err != nil {
		return nil, err
	}

	return []cache.Option{
		cache.NatsConfig(nats),
		cache.InfoConfig(c.StoreDir, c.WantMassQuery, c.Priority),
		cache.PrivateKeyFile(c.PrivateKeyFile),
	}, nil
}

func (sc *cacheRunCmd) Run() error {
	return runCmd(func(logger log.Logger, ctx context.Context) error {
		// process options
		options, err := Cli.Cache.toOptions()
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

func (sc *cacheGcCmd) Run() error {
	return runCmd(func(logger log.Logger, ctx context.Context) error {
		// process options
		options, err := Cli.Cache.toOptions()
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
