package types

import (
	"context"
	"errors"
)

var ErrUnauthenticated = errors.New("unauthenticated")

type UserKey struct{}

// SetUserKey returns a new context with the username.
func SetUserKey(ctx context.Context, username string) context.Context {
	return context.WithValue(ctx, UserKey{}, username)
}

// GetUserKey returns username from context.
func GetUserKey(ctx context.Context) (string, error) {
	userVal := ctx.Value(UserKey{})
	if userVal == nil {
		return "", ErrUnauthenticated
	}

	return userVal.(string), nil
}

type UseSeqQL struct{}

// GetUseSeqQL returns header `use-seq-ql` from context.
func GetUseSeqQL(ctx context.Context) string {
	useSeqQLRaw := ctx.Value(UseSeqQL{})
	if useSeqQLRaw == nil {
		return ""
	}
	useSeqQL, ok := useSeqQLRaw.(string)
	if !ok {
		return ""
	}
	return useSeqQL
}

type EnvKey struct{}

// SetEnvKey returns a new context with the env.
func SetEnvKey(ctx context.Context, env string) context.Context {
	return context.WithValue(ctx, EnvKey{}, env)
}

// GetEnvKey returns env from context.
func GetEnvKey(ctx context.Context) string {
	if v := ctx.Value(EnvKey{}); v != nil {
		return v.(string)
	}
	return ""
}
