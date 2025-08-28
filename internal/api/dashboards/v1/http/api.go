package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/ozontech/seq-ui/internal/api/profiles"
	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/internal/pkg/service"
)

type API struct {
	service  service.Service
	profiles *profiles.Profiles
}

func New(svc service.Service, p *profiles.Profiles) *API {
	return &API{
		service:  svc,
		profiles: p,
	}
}

func (a *API) Router() chi.Router {
	mux := chi.NewMux()

	mux.Post("/all", a.serveGetAll)
	mux.Post("/my", a.serveGetMy)
	mux.Get("/{uuid}", a.serveGetByUUID)
	mux.Post("/", a.serveCreate)
	mux.Patch("/{uuid}", a.serveUpdate)
	mux.Delete("/{uuid}", a.serveDelete)
	mux.Post("/search", a.serveSearch)

	return mux
}

type info struct {
	UUID string `json:"uuid"`
	Name string `json:"name"`
} // @name dashboards.v1.Info

func newInfo(t types.DashboardInfo) info {
	return info{
		UUID: t.UUID,
		Name: t.Name,
	}
}

type infos []info

func newInfos(t types.DashboardInfos) infos {
	res := make(infos, len(t))
	for i, d := range t {
		res[i] = newInfo(d)
	}
	return res
}

type infoWithOwner struct {
	info
	OwnerName string `json:"owner_name"`
} // @name dashboards.v1.InfoWithOwner

func newInfoWithOwner(t types.DashboardInfoWithOwner) infoWithOwner {
	return infoWithOwner{
		info:      newInfo(t.DashboardInfo),
		OwnerName: t.OwnerName,
	}
}

type infosWithOwner []infoWithOwner

func newInfosWithOwner(t types.DashboardInfosWithOwner) infosWithOwner {
	res := make(infosWithOwner, len(t))
	for i, d := range t {
		res[i] = newInfoWithOwner(d)
	}
	return res
}

type dashboard struct {
	Name      string `json:"name"`
	Meta      string `json:"meta"`
	OwnerName string `json:"owner_name"`
} // @name dashboards.v1.Dashboard

func newDashboard(d types.Dashboard) dashboard {
	return dashboard{
		Name:      d.Name,
		Meta:      d.Meta,
		OwnerName: d.OwnerName,
	}
}
