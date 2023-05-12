package cache

import (
	"bytes"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nats-io/nats.go"
	"github.com/nix-community/go-nix/pkg/narinfo"
	"go.uber.org/zap"
	"io"
	"moul.io/chizap"
	"net/http"
	"strconv"
	"time"
)

const (
	RouteCatchAll  = "/*"
	RouteNar       = "/nar/{hash:[a-z0-9]+}.nar*"
	RouteNarInfo   = "/{hash:[a-z0-9]+}.narinfo"
	RouteCacheInfo = "/nix-cache-info"

	DefaultChunkSize = int64(1024 * 1024)
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

	router.Get(RouteNarInfo, s.getNarInfo())

	router.Head(RouteNar, s.getNar(false))
	router.Get(RouteNar, s.getNar(true))

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

func (s *Cache) getNar(body bool) http.HandlerFunc {
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

		w.Header().Set("Content-Length", strconv.FormatUint(info.Size, 10))

		if !body {
			return
		}

		var written int64
		for written, err = io.CopyN(w, obj, DefaultChunkSize); written == DefaultChunkSize; {
		}
	}
}

func (s *Cache) getNarInfo() http.HandlerFunc {
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

		info, err := narinfo.Parse(obj)

		var sign = true
		for _, sig := range info.Signatures {
			if sig.Name == s.Options.Name {
				// no need to sign
				sign = false
				break
			}
		}

		if sign {
			sig, err := s.Options.SecretKey.Sign(nil, info.Fingerprint())
			if err != nil {
				s.log.Error("failed to generate nar info signature", zap.Error(err))
				w.WriteHeader(500)
				return
			}
			info.Signatures = append(info.Signatures, sig)
		}

		res := []byte(info.String())

		if sign {
			// update store
			_, err = s.store.Put(&nats.ObjectMeta{
				Name: r.RequestURI,
			}, bytes.NewReader(res))
			if err != nil {
				s.log.Error("failed to put updated nar info into NATS", zap.Error(err))
				w.WriteHeader(500)
				return
			}
		}

		w.Header().Set("Content-Length", strconv.FormatUint(uint64(len(res)), 10))
		_, err = w.Write(res)
		if err != nil {
			s.log.Error("failed to write response", zap.Error(err))
		}
	}
}
