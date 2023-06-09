package agent

import (
	"context"
	"encoding/json"
	log "github.com/inconshreveable/log15"
	"github.com/numtide/nits/pkg/nix"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/numtide/nits/pkg/cache"
	"github.com/numtide/nits/pkg/guvnor"
	"github.com/numtide/nits/pkg/state"
	"golang.org/x/sync/errgroup"
)

func (a *Agent) listenForDeployment(ctx context.Context) error {
	kv, err := state.Deployment(a.js)
	if err != nil {
		return err
	}

	agentOutput, err := state.AgentOutput(a.js)
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
				a.onDeployment(&config, agentOutput)
			}
		}
	}
}

func (a *Agent) onDeployment(config *guvnor.Deployment, agentOutput nats.KeyValue) {
	l := a.logger.New("closure", config.Closure)

	currentSystem, err := nix.CurrentSystemClosure()
	if err != nil {
		l.Error("failed to retrieve current system closure", "error", err)
	}

	if currentSystem == config.Closure {
		l.Info("current system matches deployment closure", "system", currentSystem)
		return
	}

	l.Info("deploying")

	startedAt := time.Now()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	eg, ctx := errgroup.WithContext(ctx)

	// set up the binary cache proxy
	var c *cache.Cache

	cacheLog := l.New("component", "cache")
	cacheLog.SetHandler(log.LvlFilterHandler(log.LvlError, l.GetHandler()))

	c, err = cache.NewCache(
		cacheLog,
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
		defer cancel()

		var stdOut, stdErr, output []byte

		defer func() {
			// todo handle output that is larger than 1 MB and therefore too large for the KV store
			_, err := agentOutput.Put(a.nkey, output)
			if err != nil {
				l.Error("failed to write command output to object store", "error", err)
			}

			l.Info("command output uploaded", "bucket", agentOutput.Bucket(), "name", a.nkey, "size", len(output))
		}()

		l.Info("copying from binary cache")
		if stdOut, stdErr, err = nix.CopyFromBinaryCache(c.ListenAddr(), config.Closure); err != nil {
			output = append(output, stdOut...)
			output = append(output, stdErr...)
			l.Error("failure whilst copying from binary cache")
			return err
		}

		l.Info("switching configuration")
		if stdOut, stdErr, err = nix.SwitchToConfiguration(config, a.Options.DryRun); err != nil {
			output = append(output, stdOut...)
			output = append(output, stdErr...)
			return err
		}

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
