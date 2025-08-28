package types

import (
	"github.com/ozontech/seq-ui/pkg/userprofile/v1"
)

type LogColumns struct {
	LogColumns []string `json:"log_columns"`
}

// User Profile
type UserProfile struct {
	ID                int64      `json:"_id"`
	UserName          string     `json:"user_name"`
	Timezone          string     `json:"timezone"`
	OnboardingVersion string     `json:"onboarding_version"`
	LogColumns        LogColumns `json:"log_columns"`
}

func (up UserProfile) ToProto() *userprofile.GetUserProfileResponse {
	return &userprofile.GetUserProfileResponse{
		Timezone:          up.Timezone,
		OnboardingVersion: up.OnboardingVersion,
		LogColumns:        &userprofile.LogColumns{LogColumns: up.LogColumns.LogColumns},
	}
}

type GetOrCreateUserProfileRequest struct {
	UserName string `json:"user_name"`
}

type UpdateUserProfileRequest struct {
	UserName          string      `json:"user_name"`
	Timezone          *string     `json:"timezone"`
	OnboardingVersion *string     `json:"onboarding_version"`
	LogColumns        *LogColumns `json:"log_columns"`
}

func (ur UpdateUserProfileRequest) IsEmpty() bool {
	return ur.Timezone == nil && ur.OnboardingVersion == nil && ur.LogColumns == nil
}

// Favorite Queries
type FavoriteQuery struct {
	ID           int64  `json:"id"`
	Query        string `json:"query"`
	Name         string `json:"name"`
	RelativeFrom uint64 `json:"relative_from"`
}

func (fq FavoriteQuery) ToProto() *userprofile.GetFavoriteQueriesResponse_Query {
	q := &userprofile.GetFavoriteQueriesResponse_Query{
		Id:    fq.ID,
		Query: fq.Query,
	}
	if fq.Name != "" {
		q.Name = new(string)
		*q.Name = fq.Name
	}
	if fq.RelativeFrom != 0 {
		q.RelativeFrom = new(uint64)
		*q.RelativeFrom = fq.RelativeFrom
	}
	return q
}

type FavoriteQueries []FavoriteQuery

func (fqs FavoriteQueries) ToProto() []*userprofile.GetFavoriteQueriesResponse_Query {
	queries := make([]*userprofile.GetFavoriteQueriesResponse_Query, len(fqs))

	for i, fq := range fqs {
		queries[i] = fq.ToProto()
	}

	return queries
}

type GetFavoriteQueriesRequest struct {
	ProfileID int64 `json:"profile_id"`
}

type GetOrCreateFavoriteQueryRequest struct {
	ProfileID    int64  `json:"profile_id"`
	Query        string `json:"query"`
	Name         string `json:"name"`
	RelativeFrom uint64 `json:"relative_from"`
}

type DeleteFavoriteQueryRequest struct {
	ID        int64 `json:"id"`
	ProfileID int64 `json:"profile_id"`
}
