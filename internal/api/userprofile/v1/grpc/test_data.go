package grpc

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/ozontech/seq-ui/internal/app/types"
	mock "github.com/ozontech/seq-ui/internal/pkg/service/userprofile/mock"
)

var (
	errSomethingWrong        = errors.New("something happened wrong")
	queryID           int64  = 1
	relativeFrom      uint64 = 300
	userName                 = "unnamed"
	query                    = "test"
	queryName                = "my query"
	timezone                 = "UTC"
	validTimezone            = "Europe/Moscow"
	onboardingVersion        = `{"name1": "ver1", "name2": "ver2"}`
	logColumns               = []string{"val1", "val2"}
)

func setupAPI(t *testing.T) (*API, *mock.MockService) {
	ctrl := gomock.NewController(t)
	mockedSvc := mock.NewMockService(ctrl)
	return New(mockedSvc), mockedSvc
}

func withUser(userName string) context.Context {
	return context.WithValue(context.Background(), types.UserKey{}, userName)
}
