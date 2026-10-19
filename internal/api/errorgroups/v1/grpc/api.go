package grpc

import (
	"context"

	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/internal/pkg/service/errorgroups"
	api "github.com/ozontech/seq-ui/pkg/errorgroups/v1"
)

type API struct {
	api.UnimplementedErrorGroupsServiceServer

	services map[string]errorgroups.Service
}

func New(serviceByEnv map[string]errorgroups.Service) *API {
	return &API{
		services: serviceByEnv,
	}
}

func (a *API) serviceByEnv(ctx context.Context, reqEnv *string) (errorgroups.Service, *string) {
	env := types.GetEnvKey(ctx)
	if (reqEnv == nil || *reqEnv == "") && env != "" {
		reqEnv = &env
	}
	return a.services[env], reqEnv
}

type trReq interface {
	GetTimeRange() *api.TimeRange
	GetDuration() *durationpb.Duration
}

func parseTimeRange(req trReq) *types.TimeRange {
	var timeRange *types.TimeRange
	if tr := req.GetTimeRange(); tr != nil {
		timeRange = &types.TimeRange{
			Duration: tr.Duration.AsDuration(),

			From: tr.From.AsTime(),
			To:   tr.To.AsTime(),
		}
	}
	if timeRange == nil && req.GetDuration() != nil {
		timeRange = &types.TimeRange{
			Duration: req.GetDuration().AsDuration(),
		}
	}
	return timeRange
}
