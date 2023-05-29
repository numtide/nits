package cache

import (
	"bytes"
	"context"
	"github.com/juju/errors"
	"github.com/nats-io/nats.go"
	"github.com/nix-community/go-nix/pkg/narinfo"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"strings"
	"sync/atomic"
	"time"
)

func (c *Cache) GarbageCollect(ctx context.Context, cutOff time.Time) (err error) {
	startedAt := time.Now()
	removed := atomic.Int64{}

	l := c.log.With(zap.Time("startedAt", startedAt), zap.Time("cutOff", cutOff))

	defer func() {
		elapsed := time.Now().Sub(startedAt)
		l.Info("garbage collection complete", zap.Duration("elapsed", elapsed), zap.Int64("removed", removed.Load()))
	}()

	l.Info("garbage collecting")

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(4)

	watcher, err := c.narInfo.WatchAll()
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			l.Debug("ctx completed")
			return g.Wait()
		case update, ok := <-watcher.Updates():
			if !ok {
				// watcher has been closed
				l.Debug("watcher has been closed")
				return g.Wait()
			}

			if update == nil {
				// we have reached the end of all the entries
				l.Debug("no more entries to process")
				return g.Wait()
			}

			if update.Operation() != nats.KeyValuePut {
				continue
			}

			key := update.Key()

			narInfo, err := narinfo.Parse(bytes.NewReader(update.Value()))
			if err != nil {
				return errors.Annotate(err, "failed to parse nar info")
			}

			g.Go(func() error {

				l := l.With(zap.String("key", key))
				l.Debug("checking access log")

				entry, err := c.narInfoAccess.Get(key)
				if !(err == nil || err == nats.ErrKeyNotFound) {
					return err
				}

				if err == nats.ErrKeyNotFound {
					return nil
				}

				if entry.Created().After(cutOff) {
					return nil
				}
				l.Info("entry has expired", zap.Time("lastAccessed", entry.Created()))

				// delete nar entry
				narKey := strings.ReplaceAll(narInfo.URL[4:], ".nar.", "-")
				if err = c.nar.Delete(narKey); err != nil {
					return err
				}
				l.Debug("removed nar file", zap.String("hash", narKey))

				// delete access entry
				if err = c.narInfoAccess.Delete(key); err != nil {
					return err
				}
				l.Debug("removed nar info access entry", zap.String("hash", key))

				// delete narInfo entry
				if err = c.narInfo.Delete(key); err != nil {
					return err
				}
				l.Debug("removed nar info entry", zap.String("hash", key))

				removed.Add(1)

				return nil
			})
		}
	}
}
