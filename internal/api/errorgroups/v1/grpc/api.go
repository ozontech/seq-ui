package grpc

import (
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/internal/pkg/service/errorgroups"
	api "github.com/ozontech/seq-ui/pkg/errorgroups/v1"
)

type API struct {
	api.UnimplementedErrorGroupsServiceServer

	service errorgroups.Service
}

func New(svc errorgroups.Service) *API {
	return &API{
		service: svc,
	}
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
