package service

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/ozontech/seq-ui/internal/app/types"
)

// GetAllDashboards from underlying repository.
func (s *service) GetAllDashboards(ctx context.Context, req types.GetAllDashboardsRequest) (types.DashboardInfosWithOwner, error) {
	if err := checkLimitOffset(req.Limit, req.Offset); err != nil {
		return nil, err
	}

	return s.repo.Dashboards.GetAll(ctx, req)
}

// GetMyDashboards from underlying repository.
func (s *service) GetMyDashboards(ctx context.Context, req types.GetUserDashboardsRequest) (types.DashboardInfos, error) {
	if err := checkLimitOffset(req.Limit, req.Offset); err != nil {
		return nil, err
	}

	return s.repo.Dashboards.GetMy(ctx, req)
}

// GetDashboardByUUID from underlying repository.
func (s *service) GetDashboardByUUID(ctx context.Context, id string) (types.Dashboard, error) {
	if err := checkUUID(id); err != nil {
		return types.Dashboard{}, err
	}

	return s.repo.Dashboards.GetByUUID(ctx, id)
}

// CreateDashboard in underlying repository.
func (s *service) CreateDashboard(ctx context.Context, req types.CreateDashboardRequest) (string, error) {
	if req.Name == "" {
		return "", types.NewErrInvalidRequestField("empty 'name'")
	}
	if req.Meta == "" {
		return "", types.NewErrInvalidRequestField("empty 'meta'")
	}

	return s.repo.Dashboards.Create(ctx, req)
}

// UpdateDashboard in underlying repository.
func (s *service) UpdateDashboard(ctx context.Context, req types.UpdateDashboardRequest) error {
	if err := checkUUID(req.UUID); err != nil {
		return err
	}
	if req.IsEmpty() {
		return types.ErrEmptyUpdateRequest
	}

	return s.repo.Dashboards.Update(ctx, req)
}

// DeleteDashboard in underlying repository.
func (s *service) DeleteDashboard(ctx context.Context, req types.DeleteDashboardRequest) error {
	if err := checkUUID(req.UUID); err != nil {
		return err
	}

	return s.repo.Dashboards.Delete(ctx, req)
}

// SearchDashboards in underlying repository.
func (s *service) SearchDashboards(ctx context.Context, req types.SearchDashboardsRequest) (types.DashboardInfosWithOwner, error) {
	if err := checkLimitOffset(req.Limit, req.Offset); err != nil {
		return nil, err
	}
	return s.repo.Dashboards.Search(ctx, req)
}

func checkUUID(v string) error {
	if _, err := uuid.FromString(v); err != nil {
		return types.NewErrInvalidRequestField("invalid uuid")
	}
	return nil
}

func checkLimitOffset(limit, offset int) error {
	if limit <= 0 {
		return types.NewErrInvalidRequestField("'limit' must be greater than 0")
	}
	if offset < 0 {
		return types.NewErrInvalidRequestField("'offset' must be non-negative")
	}
	return nil
}
