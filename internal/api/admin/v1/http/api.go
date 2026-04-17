package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/ozontech/seq-ui/internal/pkg/service"
)

type API struct {
	service service.Service
}

func New(svc service.Service) *API {
	return &API{
		service: svc,
	}
}

func (a *API) Router() chi.Router {
	mux := chi.NewMux()

	mux.Route("/roles", func(r chi.Router) {
		r.Post("/", a.serveCreateRole)
		r.Get("/", a.serveGetRoles)

		r.Route("/{id}", func(r chi.Router) {
			r.Post("/users", a.serveAddUsersToRole)
			r.Get("/", a.serveGetRole)
			r.Patch("/", a.serveUpdateRole)
			r.Delete("/", a.serveDeleteRole)
		})
	})

	return mux
}
