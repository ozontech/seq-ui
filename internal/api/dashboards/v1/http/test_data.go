package http

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
	"go.uber.org/mock/gomock"

	mock "github.com/ozontech/seq-ui/internal/pkg/service/dashboards/mock"
)

var (
	errSomethingWrong = errors.New("something happened wrong")
	userName          = "unnamed"
	dashboardUUID     = "064dc707-02b8-7000-8201-02a7f396738a"
	dashboardName     = "my_dashboard"
	dashboardMeta     = "my_meta"
	query             = "test-query"
	limit             = 2
	offset            = 0
)

func setupAPI(t *testing.T) (*API, *mock.MockService) {
	ctrl := gomock.NewController(t)
	mockedSvc := mock.NewMockService(ctrl)
	return New(mockedSvc), mockedSvc
}

func withUUID(h http.HandlerFunc, uuid string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rCtx := chi.NewRouteContext()
		rCtx.URLParams.Add("uuid", uuid)
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rCtx))
		h(w, r)
	}
}
