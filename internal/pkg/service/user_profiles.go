package service

import (
	"context"
	"time"

	"github.com/ozontech/seq-ui/internal/app/types"
)

// GetOrCreateUserProfile from underlying repository.
func (s *service) GetOrCreateUserProfile(ctx context.Context, req types.GetOrCreateUserProfileRequest) (types.UserProfile, error) {
	return s.repo.UserProfiles.GetOrCreate(ctx, req)
}

// UpdateUserProfile in underlying repository.
func (s *service) UpdateUserProfile(ctx context.Context, req types.UpdateUserProfileRequest) error {
	if req.IsEmpty() {
		return types.ErrEmptyUpdateRequest
	}
	if req.Timezone != nil {
		if _, err := time.LoadLocation(*req.Timezone); err != nil {
			return types.NewErrInvalidRequestField("invalid timezone format")
		}
	}

	return s.repo.UserProfiles.Update(ctx, req)
}
