package http

import (
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/ozontech/seq-ui/internal/pkg/service/errorgroups"
)

type API struct {
	service *errorgroups.Service
}

func New(service *errorgroups.Service) *API {
	return &API{
		service: service,
	}
}

func (a *API) Router() chi.Router {
	mux := chi.NewMux()

	mux.Post("/groups", a.serveGetGroups)
	mux.Post("/hist", a.serveGetHist)
	mux.Post("/details", a.serveGetDetails)
	mux.Post("/releases", a.serveGetReleases)
	mux.Post("/services", a.serveGetServices)
	mux.Post("/diff_by_releases", a.serveGetDiffByReleases)

	return mux
}

func parseGroupHash(groupHash *string) (*uint64, error) {
	if groupHash == nil {
		return nil, nil
	}

	parsedGroupHash, err := strconv.ParseUint(*groupHash, 10, 64)
	return &parsedGroupHash, err
}

func parseDuration(d *string) (*time.Duration, error) {
	if d == nil {
		return nil, nil
	}

	parsedDuration, err := time.ParseDuration(*d)
	return &parsedDuration, err
}
