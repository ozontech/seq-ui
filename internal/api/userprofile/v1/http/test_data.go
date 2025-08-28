package http

import (
	"testing"

	"github.com/ozontech/seq-ui/internal/api/userprofile/v1/test"
	repo_mock "github.com/ozontech/seq-ui/internal/pkg/repository/mock"
)

func newUserProfilesTestData(t *testing.T) (*API, *repo_mock.MockUserProfiles) {
	mock, s, p := test.NewUserProfilesData(t)
	return New(s, p), mock
}

func newFavoriteQueriesTestData(t *testing.T) (*API, *repo_mock.MockFavoriteQueries) {
	mock, s, p := test.NewFavoriteQueriesTestData(t)
	return New(s, p), mock
}
