package mw

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseURI(t *testing.T) {
	tCases := []struct {
		name        string
		uri         string
		wantApi     string
		wantVersion string
		wantMethod  string
		wantErr     bool
	}{
		{
			name:        "mid_path",
			uri:         "/api/v1/method",
			wantApi:     "api",
			wantVersion: "v1",
			wantMethod:  "method",
		},
		{
			name:        "long_path",
			uri:         "/api/v2/method/other1/other2",
			wantApi:     "api",
			wantVersion: "v2",
			wantMethod:  "method",
		},
		{
			name:    "err_empty_uri",
			uri:     "",
			wantErr: true,
		},
		{
			name:    "err_no_first_slash",
			uri:     "test/test",
			wantErr: true,
		},
		{
			name:    "err_single_slash",
			uri:     "/",
			wantErr: true,
		},
		{
			name:    "err_short_len1",
			uri:     "/method",
			wantErr: true,
		},
		{
			name:    "err_short_len2",
			uri:     "/method/",
			wantErr: true,
		},
	}

	for _, tCase := range tCases {
		tCase := tCase
		t.Run(tCase.name, func(t *testing.T) {
			t.Parallel()

			gotApi, gotVersion, gotMethod, err := parseURI(tCase.uri)
			assert.Equal(t, tCase.wantErr, err != nil)
			assert.Equal(t, tCase.wantApi, gotApi)
			assert.Equal(t, tCase.wantVersion, gotVersion)
			assert.Equal(t, tCase.wantMethod, gotMethod)
		})
	}
}

func TestParseGRPCFullMethod(t *testing.T) {
	tCases := []struct {
		name        string
		fullMethod  string
		wantService string
		wantMethod  string
		wantErr     bool
	}{
		{
			name:        "ok_with_prefix",
			fullMethod:  "/api.v1.Service/Method",
			wantService: "Service",
			wantMethod:  "Method",
		},
		{
			name:        "ok_no_prefix",
			fullMethod:  "/Service/Method",
			wantService: "Service",
			wantMethod:  "Method",
		},
		{
			name:    "empty",
			wantErr: true,
		},
		{
			name:       "no_first_slash",
			fullMethod: "api.v1.Service/Method",
			wantErr:    true,
		},
		{
			name:       "incorrect_format",
			fullMethod: "/api.v1.Service/Method/Test",
			wantErr:    true,
		},
	}

	for _, tCase := range tCases {
		tCase := tCase
		t.Run(tCase.name, func(t *testing.T) {
			t.Parallel()

			gotService, gotMethod, err := parseGRPCFullMethod(tCase.fullMethod)
			assert.Equal(t, tCase.wantErr, err != nil)
			assert.Equal(t, tCase.wantService, gotService)
			assert.Equal(t, tCase.wantMethod, gotMethod)
		})
	}
}
