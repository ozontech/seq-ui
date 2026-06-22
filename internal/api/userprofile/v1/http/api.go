package http

import (
	"github.com/go-chi/chi/v5"

	userprofile "github.com/ozontech/seq-ui/internal/pkg/service/userprofile"
)

type API struct {
	service userprofile.Service
}

func New(svc userprofile.Service) *API {
	return &API{
		service: svc,
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
