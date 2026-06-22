package grpc

import (
	"errors"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	mock_massexport "github.com/ozontech/seq-ui/internal/pkg/service/massexport/mock"
)

var (
	errSomethingWrong = errors.New("something happened wrong")
	query             = "message:error"
	testName          = "test-export"
	testWindow        = "10s"
	sessionID         = "test-session-id"
	userID            = "test-user"
	from              = time.Now()
	to                = from.Add(10 * time.Minute)
)

func setupAPI(t *testing.T) (*API, *mock_massexport.MockService) {
	ctrl := gomock.NewController(t)
	svcMock := mock_massexport.NewMockService(ctrl)
	return New(svcMock), svcMock
}
