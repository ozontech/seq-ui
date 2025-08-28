package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/ozontech/seq-ui/internal/api/profiles"
	"github.com/ozontech/seq-ui/internal/pkg/service"
)

type API struct {
	service  service.Service
	profiles *profiles.Profiles
}

func New(svc service.Service, p *profiles.Profiles) *API {
	return &API{
		service:  svc,
		profiles: p,
	}
}

func (a *API) Router() chi.Router {
	mux := chi.NewMux()

	mux.Route("/profile", func(r chi.Router) {
		r.Get("/", a.serveGetUserProfile)
		r.Patch("/", a.serveUpdateUserProfile)
	})

	mux.Route("/queries/favorite", func(r chi.Router) {
		r.Get("/", a.serveGetFavoriteQueries)
		r.Post("/", a.serveCreateFavoriteQuery)

		r.Delete("/{id}", a.serveDeleteFavoriteQuery)
	})

	return mux
}
