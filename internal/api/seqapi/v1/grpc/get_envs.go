package grpc

import (
	"context"

	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
)

func (a *API) GetEnvs(_ context.Context, _ *seqapi.GetEnvsRequest) (*seqapi.GetEnvsResponse, error) {
	return a.envsResponse, nil
}
