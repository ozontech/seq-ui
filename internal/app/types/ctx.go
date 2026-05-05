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
	userStr := ""
	userVal := ctx.Value(UserKey{})

	if userVal == nil {
		return "", ErrUnauthenticated
	}

	userStr = userVal.(string)

	return userStr, nil
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
