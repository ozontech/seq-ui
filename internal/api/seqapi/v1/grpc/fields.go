package grpc

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/logger"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
	"github.com/ozontech/seq-ui/tracing"
)

func (a *API) GetFields(ctx context.Context, req *seqapi.GetFieldsRequest) (*seqapi.GetFieldsResponse, error) {
	ctx, span := tracing.StartSpan(ctx, "seqapi_v1_get_fields")
	defer span.End()

	env := types.GetEnvKey(ctx)
	params, err := a.GetParams(env)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if cached, ok := params.caches.Fields.Get(ctx, ""); ok {
		var resp seqapi.GetFieldsResponse
		if err = proto.Unmarshal([]byte(cached), &resp); err == nil {
			return &seqapi.GetFieldsResponse{
				Fields:       resp.GetFields(),
				SystemFields: params.systemFields,
				PinnedFields: params.pinnedFields,
			}, nil
		}
		logger.Error("failed to unmarshal cached fields", zap.Error(err))
	}

	resp, err := params.client.GetFields(ctx, req)
	if err != nil {
		return nil, err
	}

	if params.caches.Fields != nil {
		if data, err := proto.Marshal(&seqapi.GetFieldsResponse{Fields: resp.GetFields()}); err == nil {
			params.caches.Fields.Set(ctx, "", string(data))
		} else {
			logger.Error("failed to marshal fields proto for caching", zap.Error(err))
		}
	}

	resp.SystemFields = params.systemFields
	resp.PinnedFields = params.pinnedFields
	return resp, nil
}

func (a *API) GetPinnedFields(ctx context.Context, _ *seqapi.GetFieldsRequest) (*seqapi.GetFieldsResponse, error) {
	env := types.GetEnvKey(ctx)
	params, err := a.GetParams(env)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &seqapi.GetFieldsResponse{
		Fields: params.pinnedFields,
	}, nil
}
