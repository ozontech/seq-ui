package http

import (
	"net/http"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/app/types"
)

func TestServeGetUserProfile(t *testing.T) {
	type mockArgs struct {
		req  types.GetOrCreateUserProfileRequest
		resp types.UserProfile
		err  error
	}

	tests := []struct {
		name string

		want    userProfile
		wantErr bool

		mockArgs *mockArgs
	}{
		{
			name: "ok",
			want: userProfile{
				Timezone:          timezone,
				OnboardingVersion: onboardingVersion,
				LogColumns:        logColumns,
			},
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
			name:    "err_svc",
			wantErr: true,
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

			httputil.DoTestHTTPEx(t, httputil.TestDataHTTPEx[struct{}, userProfile]{
				Method:  http.MethodGet,
				Target:  "/userprofile/v1/profile",
				Handler: withUser(api.serveGetUserProfile, userName),
				Want:    tt.want,
				WantErr: tt.wantErr,
			})
		})
	}
}

func TestServeUpdateUserProfile(t *testing.T) {
	type mockArgs struct {
		req types.UpdateUserProfileRequest
		err error
	}

	tests := []struct {
		name string

		req     updateUserProfileRequest
		wantErr bool

		mockArgs *mockArgs
	}{
		{
			name: "ok",
			req: updateUserProfileRequest{
				Timezone:          &validTimezone,
				OnboardingVersion: &onboardingVersion,
				LogColumns: &struct {
					Columns []string "json:\"columns\""
				}{Columns: logColumns},
			},
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
			req: updateUserProfileRequest{
				Timezone:          &validTimezone,
				OnboardingVersion: &onboardingVersion,
				LogColumns: &struct {
					Columns []string "json:\"columns\""
				}{Columns: logColumns},
			},
			wantErr: true,
			mockArgs: &mockArgs{
				req: types.UpdateUserProfileRequest{
					UserName:          userName,
					Timezone:          &validTimezone,
					OnboardingVersion: &onboardingVersion,
					LogColumns:        &types.LogColumns{LogColumns: logColumns},
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

			httputil.DoTestHTTPEx(t, httputil.TestDataHTTPEx[updateUserProfileRequest, struct{}]{
				Method:  http.MethodPatch,
				Target:  "/userprofile/v1/profile",
				Req:     tt.req,
				Handler: withUser(api.serveUpdateUserProfile, userName),
				NoResp:  true,
				WantErr: tt.wantErr,
			})
		})
	}
}
