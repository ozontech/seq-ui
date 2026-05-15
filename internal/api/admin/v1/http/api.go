package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ozontech/seq-ui/internal/app/config"
	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/internal/pkg/service"
)

type API struct {
	service              service.Service
	availablePermissions []permission
	superUsers           map[string]struct{}
}

func New(svc service.Service, cfg *config.Admin) *API {
	su := make(map[string]struct{}, len(cfg.SuperUsers))
	for _, user := range cfg.SuperUsers {
		su[user] = struct{}{}
	}

	return &API{
		service:              svc,
		availablePermissions: parsePermissions(svc.GetAvailablePermissions()),
		superUsers:           su,
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

	mux.Use(a.adminAuthInterceptor)
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
