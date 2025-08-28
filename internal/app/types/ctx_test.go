package types

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetUserKey(t *testing.T) {
	tCases := []struct {
		name         string
		ctx          context.Context
		wantUserName string
		wantErr      error
	}{
		{
			name:         "success",
			ctx:          context.WithValue(context.Background(), UserKey{}, "unnamed"),
			wantUserName: "unnamed",
		},
		{
			name:    "err_unauthenticated",
			ctx:     context.Background(),
			wantErr: ErrUnauthenticated,
		},
		{
			name:    "err_bad_value_type",
			ctx:     context.WithValue(context.Background(), UserKey{}, 99999),
			wantErr: ErrBadUserKeyValueType,
		},
	}

	for _, tCase := range tCases {
		tCase := tCase
		t.Run(tCase.name, func(t *testing.T) {
			t.Parallel()

			gotUserName, gotErr := GetUserKey(tCase.ctx)
			assert.ErrorIs(t, tCase.wantErr, gotErr)
			assert.Equal(t, tCase.wantUserName, gotUserName)
		})
	}
}
