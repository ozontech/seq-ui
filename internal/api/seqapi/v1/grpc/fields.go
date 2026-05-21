package grpc

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ozontech/seq-ui/logger"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
	"github.com/ozontech/seq-ui/tracing"
)

func (a *API) GetFields(ctx context.Context, req *seqapi.GetFieldsRequest) (*seqapi.GetFieldsResponse, error) {
	ctx, span := tracing.StartSpan(ctx, "seqapi_v1_get_fields")
	defer span.End()

	env := a.GetEnvFromContext(ctx)
	params, err := a.GetParams(env)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if params.fieldsCache == nil {
		resp, err := params.client.GetFields(ctx, req)
		if err != nil {
			return nil, err
		}

		resp.SystemFields = params.systemFields
		resp.PinnedFields = params.pinnedFields
		return resp, nil
	}

	fields, cached, isActual := params.fieldsCache.getFields()
	if cached && isActual {
		return &seqapi.GetFieldsResponse{
			Fields:       fields,
			SystemFields: params.systemFields,
			PinnedFields: params.pinnedFields,
		}, nil
	}

	resp, err := params.client.GetFields(ctx, req)
	if err != nil {
		if cached {
			logger.Error("can't get fields; use cached fields", zap.Error(err))
			return &seqapi.GetFieldsResponse{
				Fields:       fields,
				SystemFields: params.systemFields,
				PinnedFields: params.pinnedFields,
			}, nil
		}

		return nil, err
	}

	params.fieldsCache.setFields(resp.GetFields())
	resp.SystemFields = params.systemFields
	resp.PinnedFields = params.pinnedFields
	return resp, nil
}

func (a *API) GetPinnedFields(ctx context.Context, _ *seqapi.GetFieldsRequest) (*seqapi.GetFieldsResponse, error) {
	ctx, span := tracing.StartSpan(ctx, "seqapi_v1_get_fields")
	defer span.End()

	env := a.GetEnvFromContext(ctx)
	params, err := a.GetParams(env)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &seqapi.GetFieldsResponse{
		Fields: params.pinnedFields,
	}, nil
}
