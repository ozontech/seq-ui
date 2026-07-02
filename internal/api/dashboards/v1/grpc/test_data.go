package grpc

import (
	"testing"

	"github.com/ozontech/seq-ui/internal/api/dashboards/v1/test"
	repo_mock "github.com/ozontech/seq-ui/internal/pkg/repository/mock"
)

func newTestData(t *testing.T) (*API, *repo_mock.MockDashboards) {
	mock, s, p := test.NewTestData(t)
	return New(s, p), mock
}
