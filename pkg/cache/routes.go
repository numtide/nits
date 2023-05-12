package cache

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"io"
	"moul.io/chizap"
	"net/http"
	"strconv"
	"time"
)

const (
	RouteCatchAll  = "/*"
	RouteCacheInfo = "/nix-cache-info"
)

func (s *Cache) createRouter() {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(60 * time.Second))
	router.Use(chizap.New(s.log, &chizap.Opts{
		WithReferer:   true,
		WithUserAgent: true,
	}))

	router.Get(RouteCacheInfo, s.getNixCacheInfo)
	router.Put(RouteCatchAll, s.put())
	router.Head(RouteCatchAll, s.get(false))
	router.Get(RouteCatchAll, s.get(true))

	s.router = router
}

func (s *Cache) getNixCacheInfo(w http.ResponseWriter, r *http.Request) {
	if err := s.Options.Info.Write(w); err != nil {
		s.log.Error("failed to write cache info response", zap.Error(err))
	}
}

func (s *Cache) put() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		meta := &nats.ObjectMeta{
			Name: r.RequestURI,
		}
		_, err := s.store.Put(meta, r.Body)
		if err != nil {
			w.WriteHeader(500)
			_, _ = w.Write(nil)
		}
	}
}

func (s *Cache) get(body bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		obj, err := s.store.Get(r.RequestURI)

		if err == nats.ErrObjectNotFound {
			w.WriteHeader(404)
			return
		}
		if err != nil {
			w.WriteHeader(500)
			return
		}

		info, err := obj.Info()
		if err != nil {
			w.WriteHeader(500)
			return
		}

		// todo set headers
		w.Header().Set("Content-Length", strconv.FormatUint(info.Size, 10))

		if !body {
			return
		}

		chunkSize := int64(1024 * 1024)

		var written int64
		for written, err = io.CopyN(w, obj, chunkSize); written == chunkSize; {
		}
	}
}
