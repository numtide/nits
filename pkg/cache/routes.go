package cache

import (
	"bytes"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nats-io/nats.go"
	"github.com/nix-community/go-nix/pkg/narinfo"
	"go.uber.org/zap"

	"moul.io/chizap"
)

const (
	RouteCatchAll  = "/*"
	RouteNar       = "/nar/{hash:[a-z0-9]+}.nar.{compression:*}"
	RouteNarInfo   = "/{hash:[a-z0-9]+}.narinfo"
	RouteCacheInfo = "/nix-cache-info"

	ContentLength      = "Content-Length"
	ContentType        = "Content-Type"
	ContentTypeNar     = "application/x-nix-nar"
	ContentTypeNarInfo = "text/x-nix-narinfo"
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

	router.Get(RouteNarInfo, s.getNarInfo())
	router.Put(RouteNarInfo, s.putNarInfo())

	router.Head(RouteNar, s.getNar(false))
	router.Get(RouteNar, s.getNar(true))
	router.Put(RouteNar, s.putNar())

	s.router = router
}

func (s *Cache) getNixCacheInfo(w http.ResponseWriter, r *http.Request) {
	if err := s.Options.Info.Write(w); err != nil {
		s.log.Error("failed to write cache info response", zap.Error(err))
	}
}

func (s *Cache) putNar() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hash := chi.URLParam(r, "hash")
		compression := chi.URLParam(r, "compression")

		name := hash + "-" + compression
		meta := &nats.ObjectMeta{Name: name}

		_, err := s.nar.Put(meta, r.Body)
		if err != nil {
			w.WriteHeader(500)
			_, _ = w.Write(nil)
		}
	}
}

func (s *Cache) getNar(body bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hash := chi.URLParam(r, "hash")
		compression := chi.URLParam(r, "compression")

		name := hash + "-" + compression
		obj, err := s.nar.Get(name)

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

		h := w.Header()
		h.Set(ContentType, ContentTypeNar)
		h.Set(ContentLength, strconv.FormatUint(info.Size, 10))

		if !body {
			return
		}

		var written int64
		chunkSize := int64(info.Opts.ChunkSize)
		for written, err = io.CopyN(w, obj, chunkSize); written == chunkSize; {
		}
	}
}

func (s *Cache) putNarInfo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hash := chi.URLParam(r, "hash")

		value, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(500)
			_, _ = w.Write(nil)
		}
		_, err = s.narInfo.Put(hash, value)
		if err != nil {
			w.WriteHeader(500)
			_, _ = w.Write(nil)
		}
	}
}

func (s *Cache) getNarInfo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hash := chi.URLParam(r, "hash")
		obj, err := s.narInfo.Get(hash)

		if err == nats.ErrKeyNotFound {
			w.WriteHeader(404)
			return
		}
		if err != nil {
			w.WriteHeader(500)
			return
		}

		info, err := narinfo.Parse(bytes.NewReader(obj.Value()))

		sign := true
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
			_, err = s.narInfo.Put(hash, res)
			if err != nil {
				s.log.Error("failed to put updated nar info into NATS", zap.Error(err))
				w.WriteHeader(500)
				return
			}
		}

		h := w.Header()
		h.Set(ContentType, ContentTypeNarInfo)
		h.Set(ContentLength, strconv.FormatInt(int64(len(res)), 10))

		_, err = w.Write(res)
		if err != nil {
			s.log.Error("failed to write response", zap.Error(err))
		}
	}
}
