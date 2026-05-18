package http

import (
	"net/http"
	"strings"

	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/internal/pkg/service"
)

var RoutePermissions = map[string]uint64{
	"/roles": service.PermissionManageRoles,
}

func (a *API) adminAuthInterceptor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		username, err := types.GetUserKey(ctx)
		if err != nil {
			http.Error(w, "Unauthenticated", http.StatusUnauthorized)
			return
		}

		if _, ok := a.superUsers[username]; ok {
			next.ServeHTTP(w, r)
			return
		}

		requiredPermission, ok := matchRoutePermissions(r.URL.Path)
		if !ok {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		permissions, err := a.service.GetUserPermissions(ctx, types.GetUserPermissionsRequest{Username: username})
		if err != nil {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		if permissions&requiredPermission == 0 {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func matchRoutePermissions(path string) (uint64, bool) {
	path = strings.TrimPrefix(path, "/admin/v1")
	prefix := path

	if idx := strings.Index(path[1:], "/"); idx != -1 {
		prefix = path[:idx+1]
	}

	permission, ok := RoutePermissions[prefix]
	return permission, ok
}
