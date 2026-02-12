package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/tracing"
	"go.opentelemetry.io/otel/attribute"
)

// serveGetUserProfile go doc.
//
//	@Router		/userprofile/v1/profile [get]
//	@ID			userprofile_v1_getUserProfile
//	@Tags		userprofile_v1
//	@Success	200		{object}	userProfile		"A successful response"
//	@Failure	default	{object}	httputil.Error	"An unexpected error response"
//	@Security	bearer
func (a *API) serveGetUserProfile(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "userprofile_v1_get_user_profile")
	defer span.End()

	wr := httputil.NewWriter(w)

	userName, err := types.GetUserKey(ctx)
	if err != nil {
		httputil.ProcessError(wr, err)
		return
	}

	req := types.GetOrCreateUserProfileRequest{
		UserName: userName,
	}
	up, err := a.service.GetOrCreateUserProfile(ctx, req)
	if err != nil {
		httputil.ProcessError(wr, err)
		return
	}

	a.profiles.SetID(userName, up.ID)

	wr.WriteJson(newUserProfile(up))
}

// serveUpdateUserProfile go doc.
//
//	@Router		/userprofile/v1/profile [patch]
//	@ID			userprofile_v1_updateUserProfile
//	@Tags		userprofile_v1
//	@Param		body	body		updateUserProfileRequest	true	"Request body"
//	@Success	200		{object}	nil							"A successful response"
//	@Failure	default	{object}	httputil.Error				"An unexpected error response"
//	@Security	bearer
func (a *API) serveUpdateUserProfile(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "userprofile_v1_update_user_profile")
	defer span.End()

	wr := httputil.NewWriter(w)

	var httpReq updateUserProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&httpReq); err != nil {
		wr.Error(fmt.Errorf("failed to parse request: %w", err), http.StatusBadRequest)
		return
	}

	span.SetAttributes(
		attribute.KeyValue{
			Key:   "timezone",
			Value: attribute.StringValue(httpReq.GetTimezone()),
		},
		attribute.KeyValue{
			Key:   "onboarding_version",
			Value: attribute.StringValue(httpReq.GetOnboardingVersion()),
		},
		attribute.KeyValue{
			Key:   "log_columns",
			Value: attribute.StringSliceValue(httpReq.GetLogColumns()),
		},
	)

	userName, err := types.GetUserKey(ctx)
	if err != nil {
		httputil.ProcessError(wr, err)
		return
	}

	req := types.UpdateUserProfileRequest{
		UserName:          userName,
		Timezone:          httpReq.Timezone,
		OnboardingVersion: httpReq.OnboardingVersion,
	}
	if httpReq.LogColumns != nil {
		req.LogColumns = &types.LogColumns{LogColumns: httpReq.LogColumns.Columns}
	}

	if err = a.service.UpdateUserProfile(ctx, req); err != nil {
		httputil.ProcessError(wr, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type userProfile struct {
	Timezone          string   `json:"timezone"`
	OnboardingVersion string   `json:"onboardingVersion"`
	LogColumns        []string `json:"log_columns"`
} //	@name	userprofile.v1.UserProfile

func newUserProfile(t types.UserProfile) userProfile {
	return userProfile{
		Timezone:          t.Timezone,
		OnboardingVersion: t.OnboardingVersion,
		LogColumns:        t.LogColumns.LogColumns,
	}
}

type updateUserProfileRequest struct {
	Timezone          *string `json:"timezone"`
	OnboardingVersion *string `json:"onboardingVersion"`
	LogColumns        *struct {
		Columns []string `json:"columns"`
	} `json:"log_columns"`
} //	@name	userprofile.v1.UpdateUserProfileRequest

func (r updateUserProfileRequest) GetTimezone() string {
	if r.Timezone != nil {
		return *r.Timezone
	}
	return ""
}

func (r updateUserProfileRequest) GetOnboardingVersion() string {
	if r.OnboardingVersion != nil {
		return *r.OnboardingVersion
	}
	return ""
}

func (r updateUserProfileRequest) GetLogColumns() []string {
	if r.LogColumns != nil {
		return r.LogColumns.Columns
	}
	return nil
}
