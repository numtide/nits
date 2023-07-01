package cache

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-http-utils/headers"
	"github.com/numtide/nits/pkg/state"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	log "github.com/inconshreveable/log15"
	"github.com/nats-io/nats.go"
	"github.com/nix-community/go-nix/pkg/narinfo"
)

const (
	RouteNar       = "/nar/{hash:[a-z0-9]+}.nar.{compression:*}"
	RouteNarInfo   = "/{hash:[a-z0-9]+}.narinfo"
	RouteCacheInfo = "/nix-cache-info"

	ContentTypeNar     = "application/x-nix-nar"
	ContentTypeNarInfo = "text/x-nix-narinfo"
)

func (c *Cache) createRouter() *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Timeout(60 * time.Second))
	router.Use(requestLogger(c.logger))
	router.Use(middleware.Recoverer)

	router.Get(RouteCacheInfo, c.getNixCacheInfo)

	router.Head(RouteNarInfo, c.getNarInfo(false))
	router.Get(RouteNarInfo, c.getNarInfo(true))
	router.Put(RouteNarInfo, c.putNarInfo())

	router.Head(RouteNar, c.getNar(false))
	router.Get(RouteNar, c.getNar(true))
	router.Put(RouteNar, c.putNar())

	return router
}

func requestLogger(logger log.Logger) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			startedAt := time.Now()
			reqId := middleware.GetReqID(r.Context())

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			// todo improve identification of request origin e.g. nats or normal

			defer func() {
				entries := []interface{}{
					"status", ww.Status(),
					"elapsed", time.Since(startedAt),
					"from", r.RemoteAddr,
					"reqId", reqId,
				}

				switch r.Method {
				case http.MethodHead, http.MethodGet:
					entries = append(entries, "bytes", ww.BytesWritten())
				case http.MethodPost, http.MethodPut, http.MethodPatch:
					entries = append(entries, "bytes", r.ContentLength)
				}

				logger.Info(fmt.Sprintf("%s %s", r.Method, r.URL.Path), entries...)
			}()

			next.ServeHTTP(ww, r)
		}
		return http.HandlerFunc(fn)
	}
}

func (c *Cache) getNixCacheInfo(w http.ResponseWriter, r *http.Request) {
	buf := bytes.NewBuffer(nil)
	bufWriter := bufio.NewWriter(buf)
	_ = c.Options.Info.Write(bufWriter)
	// todo handle error

	_ = bufWriter.Flush()
	b := buf.Bytes()
	w.Header().Set("Content-Length", strconv.Itoa(len(b)))
	_, _ = w.Write(b)
}

func (c *Cache) putNar() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hash := chi.URLParam(r, "hash")
		compression := chi.URLParam(r, "compression")

		name := hash + "-" + compression
		meta := &nats.ObjectMeta{Name: name}

		_, err := state.Nar.Put(meta, r.Body)
		if err != nil {
			w.WriteHeader(500)
			_, _ = w.Write(nil)
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func (c *Cache) getNar(body bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hash := chi.URLParam(r, "hash")
		compression := chi.URLParam(r, "compression")

		name := hash + "-" + compression
		obj, err := state.Nar.Get(name)

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
		h.Set(headers.ContentType, ContentTypeNar)

		if !body {
			h.Set(headers.ContentLength, strconv.FormatUint(info.Size, 10))
			w.WriteHeader(http.StatusNoContent)
			return
		}

		h.Set("Transfer-Encoding", "chunked")

		written, err := io.CopyN(w, obj, int64(info.Size))
		if written != int64(info.Size) {
			log.Error("Bytes copied does not match object size", "expected", info.Size, "written", written)
		}
	}
}

func (c *Cache) putNarInfo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hash := chi.URLParam(r, "hash")

		var err error
		var info *narinfo.NarInfo

		info, err = narinfo.Parse(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("Could not parse nar info"))
		}

		sign := true
		for _, sig := range info.Signatures {
			if sig.Name == c.name {
				// no need to sign
				sign = false
				break
			}
		}

		if sign {
			sig, err := c.Options.SecretKey.Sign(nil, info.Fingerprint())
			if err != nil {
				c.logger.Error("failed to generate nar info signature", "error", err)
				w.WriteHeader(500)
				return
			}
			info.Signatures = append(info.Signatures, sig)
		}

		_, err = state.NarInfo.Put(hash, []byte(info.String()))
		if err != nil {
			w.WriteHeader(500)
			_, _ = w.Write(nil)
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func (c *Cache) getNarInfo(body bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hash := chi.URLParam(r, "hash")
		entry, err := state.NarInfo.Get(hash)

		if err == nats.ErrKeyNotFound {
			w.WriteHeader(404)
			return
		}
		if err != nil {
			w.WriteHeader(500)
			return
		}

		h := w.Header()
		h.Set(headers.ContentType, ContentTypeNarInfo)
		h.Set(headers.ContentLength, strconv.FormatInt(int64(len(entry.Value())), 10))

		if !body {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		_, err = w.Write(entry.Value())
		if err != nil {
			c.logger.Error("failed to write response", "error", err)
		}
	}
}
