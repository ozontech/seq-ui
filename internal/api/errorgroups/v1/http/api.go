package http

import (
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/internal/pkg/service/errorgroups"
)

type API struct {
	service errorgroups.Service
}

func New(svc errorgroups.Service) *API {
	return &API{
		service: svc,
	}
}

func (a *API) Router() chi.Router {
	mux := chi.NewMux()

	mux.Post("/groups", a.serveGetGroups)
	mux.Post("/top_groups", a.serveGetTopGroups)
	mux.Post("/hist", a.serveGetHist)
	mux.Post("/details", a.serveGetDetails)
	mux.Post("/releases", a.serveGetReleases)
	mux.Post("/services", a.serveGetServices)
	mux.Post("/diff_by_releases", a.serveDiffByReleases)

	return mux
}

type timeRange struct {
	Duration string `json:"duration" format:"duration" example:"1h"`

	From time.Time `json:"from" format:"date-time"`
	To   time.Time `json:"to" format:"date-time"`
} //	@name	errorgroups.v1.TimeRange

func parseTimeRange(tr *timeRange, dur *string) (*types.TimeRange, error) {
	var timeRange *types.TimeRange
	if tr != nil {
		var (
			d   time.Duration
			err error
		)
		if tr.Duration != "" {
			d, err = time.ParseDuration(tr.Duration)
			if err != nil {
				return nil, err
			}
		}

		timeRange = &types.TimeRange{
			Duration: d,

			From: tr.From,
			To:   tr.To,
		}
	}
	if timeRange == nil && dur != nil && *dur != "" {
		d, err := time.ParseDuration(*dur)
		if err != nil {
			return nil, err
		}
		timeRange = &types.TimeRange{
			Duration: d,
		}
	}
	return timeRange, nil
}

func parseGroupHash(groupHash *string) (*uint64, error) {
	if groupHash == nil {
		return nil, nil
	}

	parsedGroupHash, err := strconv.ParseUint(*groupHash, 10, 64)
	return &parsedGroupHash, err
}
