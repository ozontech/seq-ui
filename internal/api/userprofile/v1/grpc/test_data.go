package grpc

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/ozontech/seq-ui/internal/app/types"
	mock "github.com/ozontech/seq-ui/internal/pkg/service/userprofile/mock"
)

// Shared test data.
var (
	errSomethingWrong = errors.New("something happened wrong")
)

func setupTestAPI(t *testing.T) (*API, *mock.MockService) {
	ctrl := gomock.NewController(t)
	mockedSvc := mock.NewMockService(ctrl)
	return New(mockedSvc), mockedSvc
}

func withUser(userName string) context.Context {
	return context.WithValue(context.Background(), types.UserKey{}, userName)
}
