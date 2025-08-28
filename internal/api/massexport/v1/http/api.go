package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/ozontech/seq-ui/internal/pkg/service/massexport"
)

type API struct {
	exporter massexport.Service
}

func New(exporter massexport.Service) *API {
	return &API{
		exporter: exporter,
	}
}

func (a *API) Router() chi.Router {
	mux := chi.NewMux()

	mux.Post("/start", a.serveStart)
	mux.Post("/check", a.serveCheck)
	mux.Post("/cancel", a.serveCancel)
	mux.Post("/restore", a.serveRestore)
	mux.Get("/jobs", a.serveJobs)

	return mux
}
