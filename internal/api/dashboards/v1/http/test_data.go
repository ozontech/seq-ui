package http

import (
	"errors"
	"testing"

	"github.com/ozontech/seq-ui/internal/api/dashboards/v1/test"
	repo_mock "github.com/ozontech/seq-ui/internal/pkg/repository/mock"
)

var (
	errSomethingWrong = errors.New("something happened wrong")
)

func newTestData(t *testing.T) (*API, *repo_mock.MockDashboards) {
	mock, s, p := test.NewTestData(t)
	return New(s, p), mock
}
