package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/app/types"
	"go.uber.org/mock/gomock"
)

func TestServeGetUserProfile(t *testing.T) {
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

		wantRespBody string
		wantStatus   int

		mockArgs *mockArgs
		noUser   bool
	}{
		{
			name:         "success",
			wantRespBody: `{"timezone":"UTC","onboardingVersion":"{\"name1\": \"ver1\", \"name2\": \"ver2\"}","log_columns":["val1","val2"]}`,
			wantStatus:   http.StatusOK,
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
			name:       "err_no_user",
			wantStatus: http.StatusUnauthorized,
			noUser:     true,
		},
		{
			name:       "err_repo_random",
			wantStatus: http.StatusInternalServerError,
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
			req := httptest.NewRequest(http.MethodGet, "/userprofile/v1/profile", http.NoBody)

			if tt.mockArgs != nil {
				mockedRepo.EXPECT().GetOrCreate(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.resp, tt.mockArgs.err).Times(1)
			}
			if !tt.noUser {
				req = req.WithContext(context.WithValue(req.Context(), types.UserKey{}, userName))
			}

			httputil.DoTestHTTP(t, httputil.TestDataHTTP{
				Req:          req,
				Handler:      api.serveGetUserProfile,
				WantRespBody: tt.wantRespBody,
				WantStatus:   tt.wantStatus,
			})
		})
	}
}

func TestServeUpdateUserProfile(t *testing.T) {
	userName := "unnamed"
	validTimezone := "Europe/Moscow"
	invalidTimezone := "invalid timezone"
	onboardingVersion := `{"name1": "ver1", "name2": "ver2"}`
	logColumns := []string{"val1", "val2"}

	formatReqBody := func(timezone, onboardingVersion string, logColumns []string) string {
		var sb strings.Builder
		sb.WriteString("{")
		if timezone != "" {
			sb.WriteString(fmt.Sprintf(`"timezone":%q`, timezone))
		}
		if onboardingVersion != "" {
			if sb.Len() > 1 {
				sb.WriteString(",")
			}
			sb.WriteString(fmt.Sprintf(`"onboardingVersion":%q`, onboardingVersion))
		}
		if logColumns != nil {
			if sb.Len() > 1 {
				sb.WriteString(",")
			}
			v, _ := json.Marshal(logColumns)
			sb.WriteString(fmt.Sprintf(`"log_columns":{"columns":%s}`, v))
		}
		sb.WriteString("}")
		return sb.String()
	}

	type mockArgs struct {
		req types.UpdateUserProfileRequest
		err error
	}

	tests := []struct {
		name string

		reqBody    string
		wantStatus int

		mockArgs *mockArgs
		noUser   bool
	}{
		{
			name:       "success_all",
			reqBody:    formatReqBody(validTimezone, onboardingVersion, logColumns),
			wantStatus: http.StatusOK,
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
			name:       "success_only_timezone",
			reqBody:    formatReqBody(validTimezone, "", nil),
			wantStatus: http.StatusOK,
			mockArgs: &mockArgs{
				req: types.UpdateUserProfileRequest{
					UserName: userName,
					Timezone: &validTimezone,
				},
			},
		},
		{
			name:       "success_only_onboarding_ver",
			reqBody:    formatReqBody("", onboardingVersion, nil),
			wantStatus: http.StatusOK,
			mockArgs: &mockArgs{
				req: types.UpdateUserProfileRequest{
					UserName:          userName,
					OnboardingVersion: &onboardingVersion,
				},
			},
		},
		{
			name:       "success_only_log_columns",
			reqBody:    formatReqBody("", "", logColumns),
			wantStatus: http.StatusOK,
			mockArgs: &mockArgs{
				req: types.UpdateUserProfileRequest{
					UserName:   userName,
					LogColumns: &types.LogColumns{LogColumns: logColumns},
				},
			},
		},
		{
			name:       "success_empty_log_columns",
			reqBody:    formatReqBody("", "", []string{}),
			wantStatus: http.StatusOK,
			mockArgs: &mockArgs{
				req: types.UpdateUserProfileRequest{
					UserName:   userName,
					LogColumns: &types.LogColumns{LogColumns: []string{}},
				},
			},
		},
		{
			name:       "err_invalid_request",
			reqBody:    "invalid-request",
			wantStatus: http.StatusBadRequest,
			noUser:     true,
		},
		{
			name:       "err_no_user",
			reqBody:    formatReqBody(validTimezone, "", nil),
			wantStatus: http.StatusUnauthorized,
			noUser:     true,
		},
		{
			name:       "err_svc_empty_request",
			reqBody:    `{}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "err_svc_invalid_timezone_format",
			reqBody:    formatReqBody(invalidTimezone, "", nil),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "err_repo_not_found",
			reqBody:    formatReqBody(validTimezone, "", nil),
			wantStatus: http.StatusNotFound,
			mockArgs: &mockArgs{
				req: types.UpdateUserProfileRequest{
					UserName: userName,
					Timezone: &validTimezone,
				},
				err: types.ErrNotFound,
			},
		},
		{
			name:       "err_repo_random",
			reqBody:    formatReqBody(validTimezone, "", nil),
			wantStatus: http.StatusInternalServerError,
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
			req := httptest.NewRequest(http.MethodPatch, "/userprofile/v1/profile", strings.NewReader(tt.reqBody))

			if tt.mockArgs != nil {
				mockedRepo.EXPECT().Update(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.err).Times(1)
			}
			if !tt.noUser {
				req = req.WithContext(context.WithValue(req.Context(), types.UserKey{}, userName))
			}

			httputil.DoTestHTTP(t, httputil.TestDataHTTP{
				Req:        req,
				Handler:    api.serveUpdateUserProfile,
				WantStatus: tt.wantStatus,
			})
		})
	}
}
