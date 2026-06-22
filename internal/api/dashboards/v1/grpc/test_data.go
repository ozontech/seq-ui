package grpc

import (
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/ozontech/seq-ui/internal/app/types"
	mock "github.com/ozontech/seq-ui/internal/pkg/service/dashboards/mock"
)

var (
	errSomethingWrong = errors.New("something happened wrong")
	userName          = "unnamed"
	dashboardUUID     = "064dc707-02b8-7000-8201-02a7f396738a"
	dashboardName     = "my_dashboard"
	dashboardMeta     = "my_meta"
	dashboardOwner    = "owner"
	limit             = 2
	offset            = 0
	filter            = &types.SearchDashboardsFilter{
		OwnerName: &userName,
	}
)

func setupAPI(t *testing.T) (*API, *mock.MockService) {
	ctrl := gomock.NewController(t)
	mockedSvc := mock.NewMockService(ctrl)
	return New(mockedSvc), mockedSvc
}
