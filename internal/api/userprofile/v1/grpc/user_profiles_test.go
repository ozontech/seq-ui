package grpc

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/pkg/userprofile/v1"
)

func TestGetUserProfile(t *testing.T) {
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
	}{
		{
			name: "ok",
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
			name:     "err_svc",
			wantCode: codes.Internal,
			mockArgs: &mockArgs{
				req: types.GetOrCreateUserProfileRequest{
					UserName: userName,
				},
				err: errSomethingWrong,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			api, mockedSvc := setupAPI(t)

			if tt.mockArgs != nil {
				mockedSvc.EXPECT().
					GetOrCreateUserProfile(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.resp, tt.mockArgs.err).
					Times(1)
			}

			ctx := withUser(userName)
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
	}{
		{
			name: "ok",
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
			name: "err_svc",
			req: &userprofile.UpdateUserProfileRequest{
				Timezone: &validTimezone,
			},
			wantCode: codes.Internal,
			mockArgs: &mockArgs{
				req: types.UpdateUserProfileRequest{
					UserName: userName,
					Timezone: &validTimezone,
				},
				err: errSomethingWrong,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			api, mockedSvc := setupAPI(t)

			if tt.mockArgs != nil {
				mockedSvc.EXPECT().
					UpdateUserProfile(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.err).
					Times(1)
			}

			ctx := withUser(userName)
			got, err := api.UpdateUserProfile(ctx, tt.req)

			require.Equal(t, tt.wantCode, status.Code(err))
			if tt.wantCode != codes.OK {
				return
			}

			require.Equal(t, tt.want, got)
		})
	}
}
