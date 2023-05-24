package cmd

import (
	"context"
	"os"

	"github.com/juju/errors"
	"github.com/numtide/nits/pkg/cache"
)

type cacheCmd struct {
	StoreDir       string   `env:"NITS_CACHE_STORE_DIR" default:"/nix/store"`
	WantMassQuery  bool     `env:"NITS_CACHE_WANT_MASS_QUERY" default:"true"`
	Priority       int      `env:"NITS_CACHE_PRIORITY" default:"1"`
	PrivateKeyFile *os.File `env:"NITS_CACHE_PRIVATE_KEY_FILE" default:"./key.sec"`
}

func (sc *cacheCmd) toOptions() ([]cache.Option, error) {
	c := Cli.Cache

	nats, err := natsConfig()
	if err != nil {
		return nil, err
	}

	return []cache.Option{
		cache.NatsConfig(nats),
		cache.InfoConfig(c.StoreDir, c.WantMassQuery, c.Priority),
		cache.PrivateKeyFile(c.PrivateKeyFile),
	}, nil
}

func (sc *cacheCmd) Run() error {
	return runCmd(func(ctx context.Context) error {
		// process options
		options, err := sc.toOptions()
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
