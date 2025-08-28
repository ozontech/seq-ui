package auth

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/ozontech/seq-ui/internal/app/config"
	"github.com/ozontech/seq-ui/internal/app/tls"
	"github.com/ozontech/seq-ui/internal/pkg/cache"
	"github.com/ozontech/seq-ui/logger"
	"github.com/ozontech/seq-ui/metric"
)

const (
	oidcClientTimeout  = time.Second * 5
	oidcCacheKeyPrefix = "auth_oidc"
)

type OIDCToken struct {
	UserName string
}

type OIDCProvider interface {
	Verify(ctx context.Context, token string) (OIDCToken, error)
}

type oidcProvider struct {
	verifiers []*oidc.IDTokenVerifier

	cache          cache.Cache
	cacheSecretKey string

	allowedClients []string
}

func NewOIDCProvider(ctx context.Context, cfg *config.OIDC, cacheCfg config.Cache) (OIDCProvider, error) {
	var (
		err       error
		oidcCache cache.Cache
		oidcCtx   context.Context
	)

	if cfg.CacheSecretKey != "" {
		logger.Info("initializing oidc cache")
		oidcCache, err = cache.NewInmemoryWithRedisOrInmemory(ctx, cacheCfg)
		if err != nil {
			return nil, fmt.Errorf("init oidc cache: %w", err)
		}
	}

	if !cfg.SkipVerify {
		oidcCtx, err = newHTTPContext(
			context.Background(),
			httpContextCfg{
				rootCA:        cfg.RootCA,
				caCert:        cfg.CACert,
				privateKey:    cfg.PrivateKey,
				sslSkipVerify: cfg.SSLSkipVerify,
			},
		)
		if err != nil {
			return nil, err
		}
	}

	oidcCfg := &oidc.Config{
		SkipClientIDCheck:          true,
		InsecureSkipSignatureCheck: cfg.SkipVerify,
	}
	verifiers := make([]*oidc.IDTokenVerifier, 0)
	for _, url := range cfg.AuthURLs {
		var verifier *oidc.IDTokenVerifier
		if cfg.SkipVerify {
			verifier = oidc.NewVerifier(url, &oidc.StaticKeySet{}, oidcCfg)
		} else {
			provider, err := oidc.NewProvider(oidcCtx, url)
			if err != nil {
				continue
			}
			verifier = provider.Verifier(oidcCfg)
		}

		verifiers = append(verifiers, verifier)
	}

	if len(verifiers) == 0 {
		return nil, fmt.Errorf("no valid OIDS auth urls")
	}

	return &oidcProvider{
		verifiers:      verifiers,
		cache:          oidcCache,
		cacheSecretKey: cfg.CacheSecretKey,
		allowedClients: cfg.AllowedClients,
	}, nil
}

type userClaims struct {
	Username string `json:"preferred_username"`
}

func (p *oidcProvider) Verify(ctx context.Context, token string) (OIDCToken, error) {
	var (
		err      error
		idToken  *oidc.IDToken
		cacheKey string
	)

	if p.cacheSecretKey != "" {
		cacheKey = fmt.Sprintf("%s_%s", oidcCacheKeyPrefix, hashToken(token, p.cacheSecretKey))
		if userName, err := p.cache.Get(ctx, cacheKey); err == nil {
			return OIDCToken{UserName: userName}, nil
		}
	}

	defer func(start time.Time) {
		took := time.Since(start)
		metric.AuthVerifyDuration.Observe(took.Seconds())
	}(time.Now())

	for _, v := range p.verifiers {
		idToken, err = v.Verify(ctx, token)
		if err == nil {
			break
		}
	}

	if err != nil {
		return OIDCToken{}, err
	}

	err = p.checkClients(idToken.Audience)
	if err != nil {
		return OIDCToken{}, err
	}

	oidcClaims := userClaims{}
	if err = idToken.Claims(&oidcClaims); err != nil {
		return OIDCToken{}, fmt.Errorf("failed to get user claims: %w", err)
	}
	if oidcClaims.Username == "" {
		return OIDCToken{}, errors.New("invalid claims")
	}

	if p.cacheSecretKey != "" {
		_ = p.cache.SetWithTTL(ctx, cacheKey, oidcClaims.Username, time.Until(idToken.Expiry))
	}

	return OIDCToken{
		UserName: oidcClaims.Username,
	}, err
}

func (p *oidcProvider) checkClients(clients []string) error {
	if len(p.allowedClients) == 0 {
		for _, client := range clients {
			metric.UnauthorizedClientsRequests.WithLabelValues(client).Inc()
		}
		return nil
	}

	for _, client := range clients {
		for _, allowedClient := range p.allowedClients {
			if client == allowedClient {
				return nil
			}
		}
	}

	return fmt.Errorf("no allowed client found (clients: %v; allowed clients: %v)", clients, p.allowedClients)
}

type httpContextCfg struct {
	rootCA        string
	caCert        string
	privateKey    string
	sslSkipVerify bool
}

func newHTTPContext(ctx context.Context, conf httpContextCfg) (context.Context, error) {
	hc := &http.Client{
		Timeout: oidcClientTimeout,
	}
	if conf.isZero() {
		return oidc.ClientContext(ctx, hc), nil
	}
	b := tls.NewConfigBuilder()
	if conf.sslSkipVerify {
		b.SetInsecureSkipVerify(conf.sslSkipVerify)
	} else if !conf.sslSkipVerify && (conf.rootCA != "" || conf.caCert != "" || conf.privateKey != "") {
		if conf.rootCA != "" {
			if err := b.AppendCARoot(conf.rootCA); err != nil {
				return nil, fmt.Errorf("can't append CA root: %w", err)
			}
		}
		if conf.caCert != "" || conf.privateKey != "" {
			if err := b.AppendX509KeyPair(conf.caCert, conf.privateKey); err != nil {
				return nil, fmt.Errorf("can't append key pair: %w", err)
			}
		}
	}
	hc.Transport = &http.Transport{
		TLSClientConfig: b.Build(),
	}
	return oidc.ClientContext(ctx, hc), nil
}

func (c httpContextCfg) isZero() bool {
	return c.rootCA == "" && c.caCert == "" && c.privateKey == "" && !c.sslSkipVerify
}

func hashToken(token, secret string) string {
	hash := sha512.New()
	_, _ = hash.Write([]byte(token + secret))
	return hex.EncodeToString(hash.Sum(nil))
}
