package cache

import (
	"bytes"
	"context"
	"strings"
	"sync/atomic"
	"time"

	"github.com/juju/errors"
	"github.com/nats-io/nats.go"
	"github.com/nix-community/go-nix/pkg/narinfo"
	"golang.org/x/sync/errgroup"
)

func (c *Cache) GarbageCollect(ctx context.Context, cutOff time.Time) (err error) {
	startedAt := time.Now()
	removed := atomic.Int64{}

	l := c.log.New("startedAt", startedAt, "cutOff", cutOff)

	defer func() {
		elapsed := time.Now().Sub(startedAt)
		l.Info("garbage collection complete", "elapsed", elapsed, "removed", removed.Load())
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
				l := l.New("key", key)
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
				l.Info("entry has expired", "lastAccessed", entry.Created())

				// delete nar entry
				narKey := strings.ReplaceAll(narInfo.URL[4:], ".nar.", "-")
				if err = c.nar.Delete(narKey); err != nil {
					return err
				}
				l.Debug("removed nar file", "hash", narKey)

				// delete access entry
				if err = c.narInfoAccess.Delete(key); err != nil {
					return err
				}
				l.Debug("removed nar info access entry", "hash", key)

				// delete narInfo entry
				if err = c.narInfo.Delete(key); err != nil {
					return err
				}
				l.Debug("removed nar info entry", "hash", key)

				removed.Add(1)

				return nil
			})
		}
	}
}