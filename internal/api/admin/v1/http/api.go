package http

import (
	"github.com/go-chi/chi/v5"

	"github.com/ozontech/seq-ui/internal/app/types"
	adminservice "github.com/ozontech/seq-ui/internal/pkg/service/admin"
)

type API struct {
	service              adminservice.Service
	availablePermissions []permissionGroup
}

func New(svc adminservice.Service) *API {
	return &API{
		service:              svc,
		availablePermissions: parseAvailablePermissions(svc.GetAvailablePermissions()),
	}
}

func (a *API) Router() chi.Router {
	mux := chi.NewMux()

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

func parseAvailablePermissions(source []types.PermissionGroup) []permissionGroup {
	availablePermissions := make([]permissionGroup, 0, len(source))
	for _, s := range source {
		availablePermissions = append(availablePermissions, permissionGroup{
			Group:       s.Group,
			Permissions: s.Permissions,
		})
	}
	return availablePermissions
}
