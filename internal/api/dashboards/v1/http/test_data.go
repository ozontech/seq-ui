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

// Shared test data.
var (
	errSomethingWrong = errors.New("something happened wrong")
	testLimit         = 2
	testOffset        = 0
)

func setupTestAPI(t *testing.T) (*API, *mock.MockService) {
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
