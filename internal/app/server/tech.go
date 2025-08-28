package server

import (
	"encoding/json"
	"fmt"
	"mime"
	"net"
	"net/http"
	"net/http/pprof"

	"github.com/go-chi/chi/v5"
	"github.com/ozontech/seq-ui/logger"
	"github.com/ozontech/seq-ui/swagger"
	"go.uber.org/zap"
)

func serveHealth(mux *chi.Mux) {
	mux.HandleFunc("/live", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("{}"))
	})
	mux.HandleFunc("/ready", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("{}"))
	})
}

const swaggerUIPrefix = "/docs/"
const swaggerJSONPath = "/swagger.json"

func defSwagger(host string) ([]byte, error) {
	s := swagger.GetSpec()
	s.Host = host

	def, err := json.Marshal(s)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal swagger spec: %w", err)
	}
	return def, nil
}

func serveSwaggerUI(mux *chi.Mux, apiPort string) {
	_ = mime.AddExtensionType(".svg", "image/svg+xml")

	// redirect /docs/swagger.json -> /swagger.json
	mux.Get(swaggerUIPrefix[:len(swaggerUIPrefix)-1]+swaggerJSONPath, func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, swaggerJSONPath, http.StatusMovedPermanently)
	})

	mux.HandleFunc(swaggerJSONPath, func(w http.ResponseWriter, r *http.Request) {
		if h, _, err := net.SplitHostPort(r.Host); err == nil {
			r.Host = net.JoinHostPort(h, apiPort)
		}
		logger.Info("swagger.json req", zap.String("host", r.Host))

		def, err := defSwagger(r.Host)
		if err != nil {
			logger.Error("failed to get swagger def", zap.Error(err))
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(def)
	})

	// redirect /docs -> /docs/
	mux.Get(swaggerUIPrefix[:len(swaggerUIPrefix)-1], func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, swaggerUIPrefix, http.StatusMovedPermanently)
	})

	mux.Handle(swaggerUIPrefix+"*", http.StripPrefix(swaggerUIPrefix, http.FileServer(swagger.GetUI())))
}

func servePprof(mux *chi.Mux) {
	mux.HandleFunc("/debug/pprof/*", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
}
