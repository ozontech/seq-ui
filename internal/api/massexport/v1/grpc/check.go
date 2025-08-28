package grpc

import (
	"context"
	"fmt"
	"time"

	"github.com/ozontech/seq-ui/internal/api/grpcutil"
	"github.com/ozontech/seq-ui/internal/app/types"
	massexport_v1 "github.com/ozontech/seq-ui/pkg/massexport/v1"
	"github.com/ozontech/seq-ui/tracing"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (a *API) Check(ctx context.Context, req *massexport_v1.CheckRequest) (*massexport_v1.CheckResponse, error) {
	ctx, span := tracing.StartSpan(ctx, "massexport_v1_check")
	defer span.End()

	span.SetAttributes(attribute.KeyValue{
		Key:   "session_id",
		Value: attribute.StringValue(req.GetSessionId()),
	})

	info, err := a.exporter.CheckExport(ctx, req.GetSessionId())
	if err != nil {
		return nil, grpcutil.ProcessError(err)
	}

	return convertExportInfo(info), nil
}

func convertExportStatus(s types.ExportStatus) massexport_v1.ExportStatus {
	switch s {
	case types.ExportStatusUnspecified:
		return massexport_v1.ExportStatus_EXPORT_STATUS_UNSPECIFIED
	case types.ExportStatusStart:
		return massexport_v1.ExportStatus_EXPORT_STATUS_START
	case types.ExportStatusCancel:
		return massexport_v1.ExportStatus_EXPORT_STATUS_CANCEL
	case types.ExportStatusFail:
		return massexport_v1.ExportStatus_EXPORT_STATUS_FAIL
	case types.ExportStatusFinish:
		return massexport_v1.ExportStatus_EXPORT_STATUS_FINISH
	default:
		panic(fmt.Sprintf("unknown export status: %s", s.String()))
	}
}

func convertExportInfo(info types.ExportInfo) *massexport_v1.CheckResponse {
	var duration time.Duration
	if !info.StartedAt.IsZero() && !info.FinishedAt.IsZero() {
		duration = info.FinishedAt.Sub(info.StartedAt)
	}

	return &massexport_v1.CheckResponse{
		Id:           info.ID,
		Status:       convertExportStatus(info.Status),
		Progress:     info.Progress,
		Links:        info.Links,
		UserId:       info.UserID,
		StartedAt:    timestamppb.New(info.StartedAt),
		FinishedAt:   timestamppb.New(info.FinishedAt),
		Duration:     durationpb.New(duration),
		Error:        info.Error,
		UnpackedSize: int64(info.TotalSize.Unpacked),
		PackedSize:   int64(info.TotalSize.Packed),
	}
}
