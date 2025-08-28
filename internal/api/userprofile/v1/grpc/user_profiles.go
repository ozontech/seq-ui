package grpc

import (
	"context"

	"github.com/ozontech/seq-ui/internal/api/grpcutil"
	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/pkg/userprofile/v1"
	"github.com/ozontech/seq-ui/tracing"
	"go.opentelemetry.io/otel/attribute"
)

// GetUserProfile returns user's profile.
func (a *API) GetUserProfile(ctx context.Context, _ *userprofile.GetUserProfileRequest) (*userprofile.GetUserProfileResponse, error) {
	ctx, span := tracing.StartSpan(ctx, "userprofile_v1_get_user_profile")
	defer span.End()

	userName, err := types.GetUserKey(ctx)
	if err != nil {
		return nil, grpcutil.ProcessError(err)
	}

	request := types.GetOrCreateUserProfileRequest{
		UserName: userName,
	}
	userProfile, err := a.service.GetOrCreateUserProfile(ctx, request)
	if err != nil {
		return nil, grpcutil.ProcessError(err)
	}

	a.profiles.SetID(userName, userProfile.ID)

	return userProfile.ToProto(), nil
}

// UpdateUserProfile updates user's profile.
func (a *API) UpdateUserProfile(ctx context.Context, req *userprofile.UpdateUserProfileRequest) (*userprofile.UpdateUserProfileResponse, error) {
	ctx, span := tracing.StartSpan(ctx, "userprofile_v1_update_user_profile")
	defer span.End()

	span.SetAttributes(
		attribute.KeyValue{
			Key:   "timezone",
			Value: attribute.StringValue(req.GetTimezone()),
		},
		attribute.KeyValue{
			Key:   "onboarding_version",
			Value: attribute.StringValue(req.GetOnboardingVersion()),
		},
	)

	userName, err := types.GetUserKey(ctx)
	if err != nil {
		return nil, grpcutil.ProcessError(err)
	}

	request := types.UpdateUserProfileRequest{
		UserName:          userName,
		Timezone:          req.Timezone,
		OnboardingVersion: req.OnboardingVersion,
	}
	if req.GetLogColumns() != nil {
		request.LogColumns = &types.LogColumns{LogColumns: req.GetLogColumns().GetLogColumns()}
	}

	if err = a.service.UpdateUserProfile(ctx, request); err != nil {
		return nil, grpcutil.ProcessError(err)
	}

	return &userprofile.UpdateUserProfileResponse{}, nil
}
