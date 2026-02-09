package http

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/logger"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
	"github.com/ozontech/seq-ui/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

// serveGetEvent go doc.
//
//	@Router		/seqapi/v1/events/{id} [get]
//	@ID			seqapi_v1_getEvent
//	@Tags		seqapi_v1
//	@Param		env		query		string				true	"Environment"
//	@Param		id		path		string				true	"Event ID"
//	@Success	200		{object}	getEventResponse	"A successful response"
//	@Failure	default	{object}	httputil.Error		"An unexpected error response"
//	@Security	bearer
func (a *API) serveGetEvent(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "seqapi_v1_get_event")
	defer span.End()

	wr := httputil.NewWriter(w)

	id := chi.URLParam(r, "id")

	span.SetAttributes(attribute.KeyValue{
		Key: "id", Value: attribute.StringValue(id),
	})

	if cached, err := a.inmemWithRedisCache.Get(ctx, id); err == nil {
		e := &seqapi.Event{}
		if err = proto.Unmarshal([]byte(cached), e); err == nil {
			if a.masker != nil {
				a.masker.Mask(e.Data)
			}
			wr.WriteJson(getEventResponseFromProto(&seqapi.GetEventResponse{Event: e}))
			return
		}
		logger.Error("failed to unmarshal cached event proto", zap.String("id", id), zap.Error(err))
	}

	resp, err := a.seqDB.GetEvent(ctx, &seqapi.GetEventRequest{
		Id: id,
	})
	if err != nil {
		wr.Error(err, http.StatusInternalServerError)
		return
	}

	if a.masker != nil && resp.Event != nil {
		a.masker.Mask(resp.Event.Data)
	}

	if data, err := proto.Marshal(resp.Event); err == nil {
		_ = a.inmemWithRedisCache.SetWithTTL(ctx, id, string(data), a.config.EventsCacheTTL)
	} else {
		logger.Error("failed to marshal event proto for caching", zap.String("id", id), zap.Error(err))
	}

	eventResp := getEventResponseFromProto(resp)

	wr.WriteJson(eventResp)
}

type event struct {
	ID   string            `json:"id"`
	Data map[string]string `json:"data" swaggertype:"object,string"`
	Time time.Time         `json:"time" format:"date-time"`
} //	@name	seqapi.v1.Event

func eventFromProto(p *seqapi.Event) event {
	data := p.GetData()
	if data == nil {
		data = make(map[string]string)
	}
	return event{
		ID:   p.GetId(),
		Data: data,
		Time: p.GetTime().AsTime(),
	}
}

type events []event

func eventsFromProto(p []*seqapi.Event) events {
	res := make(events, len(p))
	for i, e := range p {
		res[i] = eventFromProto(e)
	}
	return res
}

type getEventResponse struct {
	Event event `json:"event"`
} //	@name	seqapi.v1.GetEventResponse

func getEventResponseFromProto(p *seqapi.GetEventResponse) getEventResponse {
	return getEventResponse{
		Event: eventFromProto(p.GetEvent()),
	}
}
