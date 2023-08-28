package server

import (
	"context"
	"errors"
	"net"

	"github.com/nats-io/nats.go"

	natshttp "github.com/brianmcgee/nats-http"
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

	return cmd.Run(logger, func(ctx context.Context) (err error) {
		var opts []nats.Option
		var conn *nats.Conn

		if opts, _, _, err = Cmd.Nats.ToOpts(); err != nil {
			return
		} else if conn, err = nats.Connect(Cmd.Nats.Url, opts...); err != nil {
			return
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

		if err = eg.Wait(); errors.Is(err, context.Canceled) {
			err = nil
		}

		return err
	})
}
