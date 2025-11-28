package seqdb

import (
	"context"
	"time"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/pkg/mask"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
)

type Client interface {
	GetAggregation(context.Context, *seqapi.GetAggregationRequest) (*seqapi.GetAggregationResponse, error)
	GetEvent(context.Context, *seqapi.GetEventRequest) (*seqapi.GetEventResponse, error)
	GetFields(context.Context, *seqapi.GetFieldsRequest) (*seqapi.GetFieldsResponse, error)
	GetHistogram(context.Context, *seqapi.GetHistogramRequest) (*seqapi.GetHistogramResponse, error)
	Search(context.Context, *seqapi.SearchRequest) (*seqapi.SearchResponse, error)
	Status(context.Context, *seqapi.StatusRequest) (*seqapi.StatusResponse, error)
	Export(context.Context, *seqapi.ExportRequest, *httputil.ChunkedWriter) error
	StartAsyncSearch(context.Context, *seqapi.StartAsyncSearchRequest) (*seqapi.StartAsyncSearchResponse, error)
	FetchAsyncSearchResult(context.Context, *seqapi.FetchAsyncSearchResultRequest) (*seqapi.FetchAsyncSearchResultResponse, error)
	GetAsyncSearchesList(context.Context, *seqapi.GetAsyncSearchesListRequest, []string) (*seqapi.GetAsyncSearchesListResponse, error)
	CancelAsyncSearch(context.Context, *seqapi.CancelAsyncSearchRequest) (*seqapi.CancelAsyncSearchResponse, error)
	DeleteAsyncSearch(context.Context, *seqapi.DeleteAsyncSearchRequest) (*seqapi.DeleteAsyncSearchResponse, error)

	// masking
	WithMasking(m *mask.Masker)
}

type GRPCKeepaliveParams struct {
	Time                time.Duration
	Timeout             time.Duration
	PermitWithoutStream bool
}

type ClientParams struct {
	Addrs               []string
	Timeout             time.Duration
	MaxRetries          int
	InitialRetryBackoff time.Duration
	MaxRetryBackoff     time.Duration
	MaxRecvMsgSize      int
	GRPCKeepaliveParams *GRPCKeepaliveParams
}

func FieldTypeToProto(t string) seqapi.FieldType {
	v, ok := seqapi.FieldType_value[t]
	if !ok {
		return seqapi.FieldType_unknown
	}
	return seqapi.FieldType(v)
}

func FieldTypeFromProto(proto seqapi.FieldType) string {
	v, ok := seqapi.FieldType_name[int32(proto)]
	if !ok {
		return seqapi.FieldType_name[int32(seqapi.FieldType_unknown)]
	}
	return v
}
