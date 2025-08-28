// Package api SeqUI Server.
//
//	@title						SeqUI Server
//	@version					1.0
//
//	@accept						json
//	@produce					json
//
//	@securityDefinitions.apikey	bearer
//	@in							header
//	@name						Authorization
//	@description				Authentication token, prefixed by Bearer: Bearer {token}
package api

import (
	"github.com/go-chi/chi/v5"
	dashboards_v1_api "github.com/ozontech/seq-ui/internal/api/dashboards/v1"
	errorgroups_v1_api "github.com/ozontech/seq-ui/internal/api/errorgroups/v1"
	massexport_v1_api "github.com/ozontech/seq-ui/internal/api/massexport/v1"
	seqapi_v1_api "github.com/ozontech/seq-ui/internal/api/seqapi/v1"
	userprofile_v1_api "github.com/ozontech/seq-ui/internal/api/userprofile/v1"
	dashboards_v1 "github.com/ozontech/seq-ui/pkg/dashboards/v1"
	errorgroups_v1 "github.com/ozontech/seq-ui/pkg/errorgroups/v1"
	massexport_v1 "github.com/ozontech/seq-ui/pkg/massexport/v1"
	seqapi_v1 "github.com/ozontech/seq-ui/pkg/seqapi/v1"
	userprofile_v1 "github.com/ozontech/seq-ui/pkg/userprofile/v1"
	"google.golang.org/grpc"
)

// Registrar is registrar of gRPC and gRPC-Gateway handlers.
type Registrar struct {
	seqApiV1      *seqapi_v1_api.SeqAPI
	userProfileV1 *userprofile_v1_api.UserProfile
	dashboardsV1  *dashboards_v1_api.Dashboards
	massExportV1  *massexport_v1_api.MassExport
	errorGroupsV1 *errorgroups_v1_api.ErrorGroups
}

// NewRegistrar returns new registrar instance.
func NewRegistrar(
	seqApiV1 *seqapi_v1_api.SeqAPI,
	userProfileV1 *userprofile_v1_api.UserProfile,
	dashboardsV1 *dashboards_v1_api.Dashboards,
	massExportV1 *massexport_v1_api.MassExport,
	errorGroupsV1 *errorgroups_v1_api.ErrorGroups,
) *Registrar {
	return &Registrar{
		seqApiV1:      seqApiV1,
		userProfileV1: userProfileV1,
		dashboardsV1:  dashboardsV1,
		massExportV1:  massExportV1,
		errorGroupsV1: errorGroupsV1,
	}
}

// RegisterGRPCHandlers registers all handlers for grpcServer.
func (r *Registrar) RegisterGRPCHandlers(grpcServer *grpc.Server) {
	if r.seqApiV1 != nil {
		seqapi_v1.RegisterSeqAPIServiceServer(grpcServer, r.seqApiV1.GRPCServer())
	}
	if r.userProfileV1 != nil {
		userprofile_v1.RegisterUserProfileServiceServer(grpcServer, r.userProfileV1.GRPCServer())
	}
	if r.dashboardsV1 != nil {
		dashboards_v1.RegisterDashboardsServiceServer(grpcServer, r.dashboardsV1.GRPCServer())
	}
	if r.massExportV1 != nil {
		massexport_v1.RegisterMassExportServiceServer(grpcServer, r.massExportV1.GRPCServer())
	}
	if r.errorGroupsV1 != nil {
		errorgroups_v1.RegisterErrorGroupsServiceServer(grpcServer, r.errorGroupsV1.GRPCServer())
	}
}

// RegisterHTTPHandlers registers all handlers for mux.
func (r *Registrar) RegisterHTTPHandlers(mux *chi.Mux) {
	if r.seqApiV1 != nil {
		mux.Mount("/seqapi/v1", r.seqApiV1.HTTPRouter())
	}
	if r.userProfileV1 != nil {
		mux.Mount("/userprofile/v1", r.userProfileV1.HTTPRouter())
	}
	if r.dashboardsV1 != nil {
		mux.Mount("/dashboards/v1", r.dashboardsV1.HTTPRouter())
	}
	if r.massExportV1 != nil {
		mux.Mount("/massexport/v1", r.massExportV1.HTTPRouter())
	}
	if r.errorGroupsV1 != nil {
		mux.Mount("/errorgroups/v1", r.errorGroupsV1.HTTPRouter())
	}
}
