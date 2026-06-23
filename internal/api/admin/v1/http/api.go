package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/ozontech/seq-ui/internal/app/types"
	adminservice "github.com/ozontech/seq-ui/internal/pkg/service/admin"
)

type API struct {
	service adminservice.Service
}

func New(svc adminservice.Service) *API {
	return &API{
		service: svc,
	}
}

func (a *API) Router() chi.Router {
	mux := chi.NewMux()

	mux.Use(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := types.SetUserKey(r.Context(), "serlazarenko")
			h.ServeHTTP(w, r.WithContext(ctx))
		})
	})

	mux.Route("/roles", func(r chi.Router) {
		r.Post("/", a.serveCreateRole)
		r.Get("/", a.serveGetRoles)

		r.Route("/{id}", func(r chi.Router) {
			r.Post("/users", a.serveAddUsersToRole)
			r.Delete("/users", a.serveDeleteUsersFromRole)
			r.Get("/", a.serveGetRole)
			r.Patch("/", a.serveUpdateRole)
			r.Delete("/", a.serveDeleteRole)
		})
	})

	return mux
}
