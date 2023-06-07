package agent

import (
	"context"
	"encoding/json"
	"github.com/nats-io/nats.go"
	"github.com/numtide/nits/pkg/cache"
	"github.com/numtide/nits/pkg/guvnor"
	"github.com/numtide/nits/pkg/state"
	"golang.org/x/sync/errgroup"
	"time"
)

func (a *Agent) listenForDeployment(ctx context.Context) error {

	kv, err := state.Deployment(a.js)
	if err != nil {
		return err
	}

	// listen for deployments using our nkey
	watch, err := kv.Watch(a.nkey)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return watch.Stop()
		case entry, ok := <-watch.Updates():
			if !ok {
				// channel has been closed
				return nil
			}
			if entry == nil {
				// nothing available yet for our nkey
				continue
			}
			if entry.Operation() == nats.KeyValuePut {
				// only process puts
				var config guvnor.Deployment
				if err = json.Unmarshal(entry.Value(), &config); err != nil {
					a.logger.Error("failed to unmarshal deployment update", "error", err)
					continue
				}
				a.onDeployment(&config)
			}
		}
	}
}

func (a *Agent) onDeployment(config *guvnor.Deployment) {
	l := a.logger.New("closure", config.Closure)
	l.Info("deploying")

	var err error
	startedAt := time.Now()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	eg, ctx := errgroup.WithContext(ctx)

	// set up the binary cache proxy
	var c *cache.Cache
	c, err = cache.NewCache(
		l.New("component", "cache"),
		cache.NatsConnection(a.conn),
		cache.BindAddress("localhost:0"), // listen to a random available port on localhost
	)
	if err != nil {
		return
	}

	if err = c.Init(); err != nil {
		return
	}

	eg.Go(func() (err error) {
		if err = c.Run(ctx); err != nil {
			return err
		}
		return nil
	})

	eg.Go(func() (err error) {
		if err = copyFromBinaryCache(c.ListenAddr(), config.Closure); err != nil {
			return err
		}

		cancel()
		return nil
	})

	defer func() {
		elapsed := time.Since(startedAt)
		if err == nil {
			l.Info("deploying complete", "elapsed", elapsed)
		} else {
			l.Error("deploying error", "elapsed", elapsed, "error", err)
		}
	}()

	err = eg.Wait()
}
