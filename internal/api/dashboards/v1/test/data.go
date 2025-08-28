package test

import (
	"testing"

	"github.com/ozontech/seq-ui/internal/api/profiles"
	repo "github.com/ozontech/seq-ui/internal/pkg/repository"
	repo_mock "github.com/ozontech/seq-ui/internal/pkg/repository/mock"
	"github.com/ozontech/seq-ui/internal/pkg/service"
	"go.uber.org/mock/gomock"
)

func NewTestData(t *testing.T) (*repo_mock.MockDashboards, service.Service, *profiles.Profiles) {
	ctl := gomock.NewController(t)
	mockedRepo := repo_mock.NewMockDashboards(ctl)
	r := &repo.Repository{
		Dashboards: mockedRepo,
	}
	s := service.New(r)
	p := profiles.New(s)
	return mockedRepo, s, p
}
