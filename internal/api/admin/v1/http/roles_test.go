package http

import (
	"errors"
	"testing"

	"github.com/ozontech/seq-ui/internal/app/types"
	mock "github.com/ozontech/seq-ui/internal/pkg/service/admin/mock"
	"go.uber.org/mock/gomock"
)

var (
	errSomethingWrong = errors.New("something happened wrong") //nolint:unused

	defaultAvailablePermissions = []types.Permission{ //nolint:unused
		{
			Value:       1,
			Name:        "manage_roles",
			Description: "Manage roles",
		},
	}
)

//nolint:unused
func setupAPI(t *testing.T) (*API, *mock.MockService) {
	ctrl := gomock.NewController(t)
	mockedSvc := mock.NewMockService(ctrl)

	mockedSvc.EXPECT().
		GetAvailablePermissions().
		Return(defaultAvailablePermissions).
		Times(1)

	api := New(mockedSvc)

	return api, mockedSvc
}
