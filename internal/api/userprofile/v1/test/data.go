package test

import (
	"testing"

	"github.com/ozontech/seq-ui/internal/api/profiles"
	repo "github.com/ozontech/seq-ui/internal/pkg/repository"
	repo_mock "github.com/ozontech/seq-ui/internal/pkg/repository/mock"
	"github.com/ozontech/seq-ui/internal/pkg/service"
	"go.uber.org/mock/gomock"
)

func NewUserProfilesData(t *testing.T) (*repo_mock.MockUserProfiles, service.Service, *profiles.Profiles) {
	ctl := gomock.NewController(t)
	mockedRepo := repo_mock.NewMockUserProfiles(ctl)
	r := &repo.Repository{
		UserProfiles: mockedRepo,
	}
	s := service.New(r)
	p := profiles.New(s)
	return mockedRepo, s, p
}

func NewFavoriteQueriesTestData(t *testing.T) (*repo_mock.MockFavoriteQueries, service.Service, *profiles.Profiles) {
	ctl := gomock.NewController(t)
	mockedRepo := repo_mock.NewMockFavoriteQueries(ctl)
	r := &repo.Repository{
		FavoriteQueries: mockedRepo,
	}
	s := service.New(r)
	p := profiles.New(s)
	return mockedRepo, s, p
}
