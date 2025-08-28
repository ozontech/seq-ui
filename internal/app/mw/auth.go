package mw

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/ozontech/seq-ui/internal/app/auth"
	"github.com/ozontech/seq-ui/internal/app/config"
	"github.com/ozontech/seq-ui/logger"
)

// nolint: gochecknoglobals
var errAuthProviderNotInit = errors.New("auth provider was not initialized")

// nolint: gochecknoglobals
var (
	tokenAuthServices = map[string]struct{}{
		"seqapi":            {},
		"massexport":        {},
		"SeqAPIService":     {},
		"MassExportService": {},
	}
)

const authHeaderBearerKey = "Bearer"

func getTokenFromAuthHeader(authHeader string) (string, error) {
	tokenOffset := len(authHeaderBearerKey) + 1
	if !strings.HasPrefix(authHeader, authHeaderBearerKey) || len(authHeader) <= tokenOffset {
		return "", errors.New("invalid authorization header value")
	}
	return authHeader[tokenOffset:], nil
}

type AuthProviders struct {
	JwtProvider  auth.JWTProvider
	OidcProvider auth.OIDCProvider
}

func NewAuthProviders(
	ctx context.Context,
	jwtSecretKey string, oidcCfg *config.OIDC,
	cacheCfg config.Cache,
) (AuthProviders, error) {
	authPrvds := AuthProviders{}

	if jwtSecretKey == "" && oidcCfg == nil {
		return authPrvds, nil
	}

	if jwtSecretKey != "" {
		logger.Info("initializing jwt provider")
		authPrvds.JwtProvider = auth.NewJWTProvider(jwtSecretKey)
	}

	if oidcCfg != nil {
		logger.Info("initializing oidc provider")
		var err error
		authPrvds.OidcProvider, err = auth.NewOIDCProvider(ctx, oidcCfg, cacheCfg)
		if err != nil {
			return authPrvds, fmt.Errorf("failed to init oidc provider: %w", err)
		}
	}
	return authPrvds, nil
}

func (p *AuthProviders) auth(ctx context.Context, authHeader string, checkAPIToken bool) (string, error) {
	token, err := getTokenFromAuthHeader(authHeader)
	if err != nil {
		return "", fmt.Errorf("failed to get token from auth header: %w", err)
	}
	if checkAPIToken && p.JwtProvider != nil {
		// first check if token was issued by the app's JWT provider
		jwtClaims, err := p.JwtProvider.Verify(token)
		if err == nil {
			return formatJWTServiceName(jwtClaims.Name), nil
		} else if p.OidcProvider == nil {
			return "", fmt.Errorf("failed to verify token: %w", err)
		}
	}

	if p.OidcProvider == nil {
		return "", errAuthProviderNotInit
	}

	// if JWT provider verification failed, check if the token was issued by OIDC
	oidcToken, err := p.OidcProvider.Verify(ctx, token)
	if err != nil {
		return "", fmt.Errorf("failed to verify token: %w", err)
	}

	return oidcToken.UserName, nil
}

func formatJWTServiceName(name string) string {
	// all api tokens names are prefixed with `api@`, so they have separate requests limits
	return fmt.Sprintf("api@%s", name)
}
