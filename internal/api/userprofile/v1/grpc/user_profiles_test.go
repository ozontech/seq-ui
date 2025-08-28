package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/pkg/userprofile/v1"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGetUserProfile(t *testing.T) {
	userName := "unnamed"
	timezone := "UTC"
	onboardingVersion := `{"name1": "ver1", "name2": "ver2"}`
	logColumns := []string{"val1", "val2"}

	type mockArgs struct {
		req  types.GetOrCreateUserProfileRequest
		resp types.UserProfile
		err  error
	}

	tests := []struct {
		name string

		want     *userprofile.GetUserProfileResponse
		wantCode codes.Code

		mockArgs *mockArgs
		noUser   bool
	}{
		{
			name: "success",
			want: &userprofile.GetUserProfileResponse{
				Timezone:          timezone,
				OnboardingVersion: onboardingVersion,
				LogColumns:        &userprofile.LogColumns{LogColumns: logColumns},
			},
			wantCode: codes.OK,
			mockArgs: &mockArgs{
				req: types.GetOrCreateUserProfileRequest{
					UserName: userName,
				},
				resp: types.UserProfile{
					ID:                1,
					UserName:          userName,
					Timezone:          timezone,
					OnboardingVersion: onboardingVersion,
					LogColumns:        types.LogColumns{LogColumns: logColumns},
				},
			},
		},
		{
			name:     "err_no_user",
			wantCode: codes.Unauthenticated,
			noUser:   true,
		},
		{
			name:     "err_repo_random",
			wantCode: codes.Internal,
			mockArgs: &mockArgs{
				req: types.GetOrCreateUserProfileRequest{
					UserName: userName,
				},
				err: errors.New("random repo err"),
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			api, mockedRepo := newUserProfilesTestData(t)

			if tt.mockArgs != nil {
				mockedRepo.EXPECT().GetOrCreate(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.resp, tt.mockArgs.err).Times(1)
			}

			ctx := context.Background()
			if !tt.noUser {
				ctx = context.WithValue(ctx, types.UserKey{}, userName)
			}

			got, err := api.GetUserProfile(ctx, &userprofile.GetUserProfileRequest{})

			require.Equal(t, tt.wantCode, status.Code(err))
			if tt.wantCode != codes.OK {
				return
			}

			require.Equal(t, tt.want, got)
		})
	}
}

func TestUpdateUserProfile(t *testing.T) {
	userName := "unnamed"
	validTimezone := "Europe/Moscow"
	invalidTimezone := "invalid timezone"
	onboardingVersion := `{"name1": "ver1", "name2": "ver2"}`
	logColumns := []string{"val1", "val2"}

	type mockArgs struct {
		req types.UpdateUserProfileRequest
		err error
	}

	tests := []struct {
		name string

		req      *userprofile.UpdateUserProfileRequest
		want     *userprofile.UpdateUserProfileResponse
		wantCode codes.Code

		mockArgs *mockArgs
		noUser   bool
	}{
		{
			name: "success_all",
			req: &userprofile.UpdateUserProfileRequest{
				Timezone:          &validTimezone,
				OnboardingVersion: &onboardingVersion,
				LogColumns:        &userprofile.LogColumns{LogColumns: logColumns},
			},
			want:     &userprofile.UpdateUserProfileResponse{},
			wantCode: codes.OK,
			mockArgs: &mockArgs{
				req: types.UpdateUserProfileRequest{
					UserName:          userName,
					Timezone:          &validTimezone,
					OnboardingVersion: &onboardingVersion,
					LogColumns:        &types.LogColumns{LogColumns: logColumns},
				},
			},
		},
		{
			name: "success_only_timezone",
			req: &userprofile.UpdateUserProfileRequest{
				Timezone: &validTimezone,
			},
			want:     &userprofile.UpdateUserProfileResponse{},
			wantCode: codes.OK,
			mockArgs: &mockArgs{
				req: types.UpdateUserProfileRequest{
					UserName: userName,
					Timezone: &validTimezone,
				},
			},
		},
		{
			name: "success_only_onboarding_ver",
			req: &userprofile.UpdateUserProfileRequest{
				OnboardingVersion: &onboardingVersion,
			},
			want:     &userprofile.UpdateUserProfileResponse{},
			wantCode: codes.OK,
			mockArgs: &mockArgs{
				req: types.UpdateUserProfileRequest{
					UserName:          userName,
					OnboardingVersion: &onboardingVersion,
				},
			},
		},
		{
			name: "success_only_log_columns",
			req: &userprofile.UpdateUserProfileRequest{
				LogColumns: &userprofile.LogColumns{LogColumns: logColumns},
			},
			want:     &userprofile.UpdateUserProfileResponse{},
			wantCode: codes.OK,
			mockArgs: &mockArgs{
				req: types.UpdateUserProfileRequest{
					UserName:   userName,
					LogColumns: &types.LogColumns{LogColumns: logColumns},
				},
			},
		},
		{
			name: "success_empty_log_columns",
			req: &userprofile.UpdateUserProfileRequest{
				LogColumns: &userprofile.LogColumns{LogColumns: []string{}},
			},
			want:     &userprofile.UpdateUserProfileResponse{},
			wantCode: codes.OK,
			mockArgs: &mockArgs{
				req: types.UpdateUserProfileRequest{
					UserName:   userName,
					LogColumns: &types.LogColumns{LogColumns: []string{}},
				},
			},
		},
		{
			name:     "err_no_user",
			wantCode: codes.Unauthenticated,
			noUser:   true,
		},
		{
			name:     "err_svc_empty_request",
			req:      &userprofile.UpdateUserProfileRequest{},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "err_svc_invalid_timezone_format",
			req: &userprofile.UpdateUserProfileRequest{
				Timezone: &invalidTimezone,
			},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "err_repo_not_found",
			req: &userprofile.UpdateUserProfileRequest{
				Timezone: &validTimezone,
			},
			wantCode: codes.NotFound,
			mockArgs: &mockArgs{
				req: types.UpdateUserProfileRequest{
					UserName: userName,
					Timezone: &validTimezone,
				},
				err: types.ErrNotFound,
			},
		},
		{
			name: "err_repo_random",
			req: &userprofile.UpdateUserProfileRequest{
				Timezone: &validTimezone,
			},
			wantCode: codes.Internal,
			mockArgs: &mockArgs{
				req: types.UpdateUserProfileRequest{
					UserName: userName,
					Timezone: &validTimezone,
				},
				err: errors.New("random repo err"),
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			api, mockedRepo := newUserProfilesTestData(t)

			if tt.mockArgs != nil {
				mockedRepo.EXPECT().Update(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.err).Times(1)
			}

			ctx := context.Background()
			if !tt.noUser {
				ctx = context.WithValue(ctx, types.UserKey{}, userName)
			}

			got, err := api.UpdateUserProfile(ctx, tt.req)

			require.Equal(t, tt.wantCode, status.Code(err))
			if tt.wantCode != codes.OK {
				return
			}

			require.Equal(t, tt.want, got)
		})
	}
}
