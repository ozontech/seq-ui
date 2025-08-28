package mw

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/ozontech/seq-ui/internal/app/auth"
	mock_auth "github.com/ozontech/seq-ui/internal/app/auth/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestAuthProvidersAuth(t *testing.T) {
	token := "test"
	authHeader := fmt.Sprintf("%s %s", authHeaderBearerKey, token)
	userName := "unnamed"
	oidcToken := auth.OIDCToken{
		UserName: userName,
	}
	jwtClaims := &auth.JWTClaims{
		Name: userName,
	}
	err := errors.New("some error")

	type (
		jwtMockArgs struct {
			token     string
			jwtClaims *auth.JWTClaims
			err       error
		}

		oidcMockArgs struct {
			token     string
			oidcToken auth.OIDCToken
			err       error
		}

		mockArgs struct {
			jwt  *jwtMockArgs
			oidc *oidcMockArgs
		}
	)

	tCases := []struct {
		name string

		authHeader    string
		checkAPIToken bool
		mockArgs      mockArgs

		want    string
		wantErr bool
	}{
		{
			name:       "err_invalid_authheader",
			authHeader: "invalid header",
			wantErr:    true,
		},
		{
			name:          "ok_jwt_only",
			authHeader:    authHeader,
			checkAPIToken: true,
			mockArgs: mockArgs{
				jwt: &jwtMockArgs{
					token:     token,
					jwtClaims: jwtClaims,
				},
			},
			want: formatJWTServiceName(userName),
		},
		{
			name:          "err_jwt_only",
			authHeader:    authHeader,
			checkAPIToken: true,
			mockArgs: mockArgs{
				jwt: &jwtMockArgs{
					token: token,
					err:   err,
				},
			},
			wantErr: true,
		},
		{
			name:          "ok_oidc_only",
			authHeader:    authHeader,
			checkAPIToken: true,
			mockArgs: mockArgs{
				oidc: &oidcMockArgs{
					token:     token,
					oidcToken: oidcToken,
				},
			},
			want: userName,
		},
		{
			name:          "err_oidc_only",
			authHeader:    authHeader,
			checkAPIToken: true,
			mockArgs: mockArgs{
				oidc: &oidcMockArgs{
					token: token,
					err:   err,
				},
			},
			wantErr: true,
		},
		{
			name:          "err_no_providers",
			authHeader:    authHeader,
			checkAPIToken: true,
			wantErr:       true,
		},
		{
			name:          "err_jwt_ok_oidc",
			authHeader:    authHeader,
			checkAPIToken: true,
			mockArgs: mockArgs{
				jwt: &jwtMockArgs{
					token: token,
					err:   err,
				},
				oidc: &oidcMockArgs{
					token:     token,
					oidcToken: oidcToken,
				},
			},
			want: userName,
		},
		{
			name:          "ok_oidc_skip_check_api_token",
			authHeader:    authHeader,
			checkAPIToken: false,
			mockArgs: mockArgs{
				jwt: &jwtMockArgs{
					token:     token,
					jwtClaims: jwtClaims,
				},
				oidc: &oidcMockArgs{
					token:     token,
					oidcToken: oidcToken,
				},
			},
			want: userName,
		},
		{
			name:          "err_oidc_skip_check_api_token",
			authHeader:    authHeader,
			checkAPIToken: false,
			mockArgs: mockArgs{
				jwt: &jwtMockArgs{
					token:     token,
					jwtClaims: jwtClaims,
				},
				oidc: &oidcMockArgs{
					token: token,
					err:   err,
				},
			},
			wantErr: true,
		},
	}

	for _, tCase := range tCases {
		tCase := tCase
		t.Run(tCase.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			authPrv := &AuthProviders{}
			if tCase.mockArgs.jwt != nil {
				jwtProvider := mock_auth.NewMockJWTProvider(ctrl)
				if tCase.checkAPIToken {
					jwtProvider.EXPECT().Verify(tCase.mockArgs.jwt.token).
						Return(tCase.mockArgs.jwt.jwtClaims, tCase.mockArgs.jwt.err)
				}

				authPrv.JwtProvider = jwtProvider
			}
			if tCase.mockArgs.oidc != nil {
				oidcProvider := mock_auth.NewMockOIDCProvider(ctrl)
				oidcProvider.EXPECT().Verify(gomock.Any(), tCase.mockArgs.oidc.token).
					Return(tCase.mockArgs.oidc.oidcToken, tCase.mockArgs.oidc.err)

				authPrv.OidcProvider = oidcProvider
			}

			got, err := authPrv.auth(context.Background(), authHeader, tCase.checkAPIToken)
			require.Equal(t, tCase.wantErr, err != nil)
			if tCase.wantErr {
				return
			}

			require.Equal(t, got, tCase.want)
		})
	}
}
