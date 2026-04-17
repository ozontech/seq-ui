package http

import (
	"net/http"

	"github.com/ozontech/seq-ui/internal/api/httputil"
)

func (a *API) serveCreateRole(w http.ResponseWriter, r *http.Request) {
	httputil.NewWriter(w).WriteHeader(http.StatusNotImplemented)
}

func (a *API) serveGetRoles(w http.ResponseWriter, r *http.Request) {
	httputil.NewWriter(w).WriteHeader(http.StatusNotImplemented)
}

func (a *API) serveAddUsersToRole(w http.ResponseWriter, r *http.Request) {
	httputil.NewWriter(w).WriteHeader(http.StatusNotImplemented)
}

func (a *API) serveGetRole(w http.ResponseWriter, r *http.Request) {
	httputil.NewWriter(w).WriteHeader(http.StatusNotImplemented)
}

func (a *API) serveUpdateRole(w http.ResponseWriter, r *http.Request) {
	httputil.NewWriter(w).WriteHeader(http.StatusNotImplemented)
}

func (a *API) serveDeleteRole(w http.ResponseWriter, r *http.Request) {
	httputil.NewWriter(w).WriteHeader(http.StatusNotImplemented)
}
