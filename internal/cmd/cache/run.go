package server

import (
	"context"
	"net"

	natshttp "github.com/brianmcgee/nats-http"
	"github.com/juju/errors"
	"github.com/numtide/nits/pkg/cache"
	"golang.org/x/sync/errgroup"

	"github.com/numtide/nits/internal/cmd"
)

type runCmd struct {
	CacheAddress string `env:"NITS_CACHE_ADDRESS" default:"localhost:3000"`
}

func (r *runCmd) Run() error {
	logger, err := Cmd.Logging.ToLogger()
	if err != nil {
		return err
	}

	if Cmd.Nats.InboxFormat == "" {
		Cmd.Nats.InboxFormat = "_INBOX.cache.%s"
	}

	return cmd.Run(logger, func(ctx context.Context) error {
		natsConfig, err := Cmd.Nats.ToNatsConfig()
		if err != nil {
			return err
		}

		conn, nkey, err := natsConfig.Connect(logger)
		if err != nil {
			return errors.Annotatef(err, "nkey = "+nkey)
		}

		cacheOptions, err := Cmd.Cache.ToCacheOptions()
		if err != nil {
			return err
		}

		c := cache.Cache{
			Conn:    conn,
			Options: *cacheOptions,
		}

		// create a http proxy for the cache service
		listener, err := net.Listen("tcp", r.CacheAddress)
		if err != nil {
			return err
		}

		proxy := natshttp.Proxy{
			Subject:  c.Options.Subject,
			Listener: listener,
			Transport: &natshttp.Transport{
				Conn: conn,
				// increase the subscription pending msg bytes to 512 MB
				PendingBytesLimit: 1024 * 1024 * 512,
			},
		}

		// run services in an error group
		eg := errgroup.Group{}

		eg.Go(func() error {
			return c.Listen(ctx, logger)
		})

		eg.Go(func() error {
			return proxy.Listen(ctx)
		})

		err = eg.Wait()
		if err == context.Canceled {
			err = nil
		}

		return err
	})
}
