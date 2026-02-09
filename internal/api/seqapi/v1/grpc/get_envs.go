package grpc

import (
	"context"

	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
)

func (a *API) GetEnvs(ctx context.Context, req *seqapi.GetEnvsRequest) (*seqapi.GetEnvsResponse, error) {
	return &seqapi.GetEnvsResponse{}, nil
}
