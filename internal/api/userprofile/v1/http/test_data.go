package http

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
	"go.uber.org/mock/gomock"

	"github.com/ozontech/seq-ui/internal/app/types"
	mock "github.com/ozontech/seq-ui/internal/pkg/service/userprofile/mock"
)

var (
	errSomethingWrong = errors.New("something happened wrong")
	userName          = "unnamed"
	relativeFrom      = "300"
	query             = "test"
	queryName         = "my query"
	timezone          = "UTC"
	validTimezone     = "Europe/Moscow"
	onboardingVersion = `{"name1": "ver1", "name2": "ver2"}`
	logColumns        = []string{"val1", "val2"}
)

func setupAPI(t *testing.T) (*API, *mock.MockService) {
	ctrl := gomock.NewController(t)
	mockedSvc := mock.NewMockService(ctrl)
	return New(mockedSvc), mockedSvc
}

func withUser(h http.HandlerFunc, userName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r = r.WithContext(context.WithValue(r.Context(), types.UserKey{}, userName))
		h(w, r)
	}
}

func withID(h http.HandlerFunc, id string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rCtx := chi.NewRouteContext()
		rCtx.URLParams.Add("id", id)
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rCtx))
		h(w, r)
	}
}
