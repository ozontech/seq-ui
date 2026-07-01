package userprofile_v1

import (
	"github.com/go-chi/chi/v5"

	grpc_api "github.com/ozontech/seq-ui/internal/api/userprofile/v1/grpc"
	http_api "github.com/ozontech/seq-ui/internal/api/userprofile/v1/http"
	"github.com/ozontech/seq-ui/internal/pkg/service/userprofile"
)

type UserProfile struct {
	grpcAPI *grpc_api.API
	httpAPI *http_api.API
}

func New(svc userprofile.Service) *UserProfile {
	return &UserProfile{
		grpcAPI: grpc_api.New(svc),
		httpAPI: http_api.New(svc),
	}
}

func (up *UserProfile) GRPCServer() *grpc_api.API {
	return up.grpcAPI
}

func (up *UserProfile) HTTPRouter() chi.Router {
	return up.httpAPI.Router()
}
