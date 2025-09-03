package types

import (
	"context"
	"errors"
)

var (
	ErrUnauthenticated     = errors.New("unauthenticated")
	ErrBadUserKeyValueType = errors.New("invalid JWT token")
)

type UserKey struct{}

// GetUserKey returns username from context.
func GetUserKey(ctx context.Context) (string, error) {
	userStr := ""
	userVal := ctx.Value(UserKey{})

	if userVal == nil {
		return "", ErrUnauthenticated
	}

	userStr, ok := userVal.(string)
	// foolproof.
	if !ok {
		return "", ErrBadUserKeyValueType
	}

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
