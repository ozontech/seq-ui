package dashboards

import (
	"context"

	"github.com/gofrs/uuid"

	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/internal/pkg/repository"
	"github.com/ozontech/seq-ui/internal/pkg/service/profiles"
)

type Service interface {
	GetAllDashboards(context.Context, types.GetAllDashboardsRequest) (types.DashboardInfosWithOwner, error)
	GetMyDashboards(context.Context, types.GetUserDashboardsRequest) (types.DashboardInfos, error)
	GetDashboardByUUID(context.Context, string) (types.Dashboard, error)
	CreateDashboard(context.Context, types.CreateDashboardRequest) (string, error)
	UpdateDashboard(context.Context, types.UpdateDashboardRequest) error
	DeleteDashboard(context.Context, types.DeleteDashboardRequest) error
	SearchDashboards(context.Context, types.SearchDashboardsRequest) (types.DashboardInfosWithOwner, error)
}

type service struct {
	repo repository.Dashboards
}

func New(repo repository.Dashboards) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) GetAllDashboards(ctx context.Context, req types.GetAllDashboardsRequest) (types.DashboardInfosWithOwner, error) {
	// check auth and create profile if its doesn't exist
	if _, err := profiles.GetIDFromContext(ctx); err != nil {
		return nil, err
	}

	if err := checkLimitOffset(req.Limit, req.Offset); err != nil {
		return nil, err
	}

	return s.repo.GetAll(ctx, req)
}

func (s *service) GetMyDashboards(ctx context.Context, req types.GetUserDashboardsRequest) (types.DashboardInfos, error) {
	profileID, err := profiles.GetIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	req.ProfileID = profileID

	if err := checkLimitOffset(req.Limit, req.Offset); err != nil {
		return nil, err
	}

	return s.repo.GetMy(ctx, req)
}

func (s *service) GetDashboardByUUID(ctx context.Context, id string) (types.Dashboard, error) {
	// check auth and create profile if its doesn't exist
	if _, err := profiles.GetIDFromContext(ctx); err != nil {
		return types.Dashboard{}, err
	}

	if err := checkUUID(id); err != nil {
		return types.Dashboard{}, err
	}

	return s.repo.GetByUUID(ctx, id)
}

func (s *service) CreateDashboard(ctx context.Context, req types.CreateDashboardRequest) (string, error) {
	profileID, err := profiles.GetIDFromContext(ctx)
	if err != nil {
		return "", err
	}
	req.ProfileID = profileID

	if req.Name == "" {
		return "", types.NewErrInvalidRequestField("empty 'name'")
	}
	if req.Meta == "" {
		return "", types.NewErrInvalidRequestField("empty 'meta'")
	}

	return s.repo.Create(ctx, req)
}

func (s *service) UpdateDashboard(ctx context.Context, req types.UpdateDashboardRequest) error {
	profileID, err := profiles.GetIDFromContext(ctx)
	if err != nil {
		return err
	}
	req.ProfileID = profileID

	if err := checkUUID(req.UUID); err != nil {
		return err
	}
	if req.IsEmpty() {
		return types.ErrEmptyUpdateRequest
	}

	return s.repo.Update(ctx, req)
}

func (s *service) DeleteDashboard(ctx context.Context, req types.DeleteDashboardRequest) error {
	profileID, err := profiles.GetIDFromContext(ctx)
	if err != nil {
		return err
	}
	req.ProfileID = profileID

	if err := checkUUID(req.UUID); err != nil {
		return err
	}

	return s.repo.Delete(ctx, req)
}

func (s *service) SearchDashboards(ctx context.Context, req types.SearchDashboardsRequest) (types.DashboardInfosWithOwner, error) {
	// check auth and create profile if its doesn't exist
	if _, err := profiles.GetIDFromContext(ctx); err != nil {
		return nil, err
	}

	if err := checkLimitOffset(req.Limit, req.Offset); err != nil {
		return nil, err
	}

	return s.repo.Search(ctx, req)
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
