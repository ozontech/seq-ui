package http

import (
	"github.com/go-chi/chi/v5"

	"github.com/ozontech/seq-ui/internal/pkg/service/massexport"
)

type API struct {
	serviceByEnv map[string]massexport.Service
}

func New(serviceByEnv map[string]massexport.Service) *API {
	return &API{
		serviceByEnv: serviceByEnv,
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
