package grpc

import (
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	mock "github.com/ozontech/seq-ui/internal/pkg/service/dashboards/mock"
)

// Shared test data.
var (
	errSomethingWrong = errors.New("something happened wrong")
	testDashboardUUID = "064dc707-02b8-7000-8201-02a7f396738a"
	testDashboardName = "my_dashboard"
	testDashboardMeta = "my_meta"
	testLimit         = 2
	testOffset        = 0
)

func setupTestAPI(t *testing.T) (*API, *mock.MockService) {
	ctrl := gomock.NewController(t)
	mockedSvc := mock.NewMockService(ctrl)
	return New(mockedSvc), mockedSvc
}
