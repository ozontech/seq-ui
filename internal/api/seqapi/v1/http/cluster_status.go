package http

import (
	"net/http"
	"time"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
	"github.com/ozontech/seq-ui/tracing"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// serveStatus go doc.
//
//	@Router		/seqapi/v1/status [get]
//	@ID			seqapi_v1_status
//	@Tags		seqapi_v1
//	@Success	200		{object}	statusResponse	"A successful response"
//	@Failure	default	{object}	httputil.Error	"An unexpected error response"
func (a *API) serveStatus(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "seqapi_v1_status")
	defer span.End()

	wr := httputil.NewWriter(w)

	resp, err := a.seqDB.Status(ctx, &seqapi.StatusRequest{})
	if err != nil {
		wr.Error(err, http.StatusInternalServerError)
		return
	}

	wr.WriteJson(getStatusResponseFromProto(resp))
}

type statusResponse struct {
	OldestStorageTime *time.Time    `json:"oldest_storage_time,omitempty"`
	NumberOfStores    int32         `json:"number_of_stores"`
	Stores            []storeStatus `json:"stores"`
} //	@name	seqapi.v1.StatusResponse

func getStatusResponseFromProto(proto *seqapi.StatusResponse) statusResponse {
	stores := make([]storeStatus, len(proto.Stores))
	for i := range stores {
		stores[i] = getStoreStatusFromProto(proto.Stores[i])
	}

	return statusResponse{
		OldestStorageTime: convertOptionalTimestamp(proto.OldestStorageTime),
		NumberOfStores:    proto.NumberOfStores,
		Stores:            stores,
	}
}

func convertOptionalTimestamp(t *timestamppb.Timestamp) *time.Time {
	if t == nil {
		return nil
	}

	result := t.AsTime()
	return &result
}

type storeStatus struct {
	Host   string             `json:"host"`
	Values *storeStatusValues `json:"values,omitempty"`
	Error  *string            `json:"error,omitempty"`
} //	@name	seqapi.v1.StoreStatus

func getStoreStatusFromProto(proto *seqapi.StoreStatus) storeStatus {
	return storeStatus{
		Host:   proto.Host,
		Values: getStoreStatusValuesFromProto(proto.Values),
		Error:  proto.Error,
	}
}

type storeStatusValues struct {
	OldestTime *time.Time `json:"oldest_time"`
} //	@name	seqapi.v1.StoreStatusValues

func getStoreStatusValuesFromProto(proto *seqapi.StoreStatusValues) *storeStatusValues {
	if proto == nil {
		return nil
	}

	oldestTime := proto.OldestTime.AsTime()
	return &storeStatusValues{OldestTime: &oldestTime}
}
