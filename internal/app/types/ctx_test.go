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
			ctx:          SetUserKey(context.Background(), "unnamed"),
			wantUserName: "unnamed",
		},
		{
			name:    "err_unauthenticated",
			ctx:     context.Background(),
			wantErr: ErrUnauthenticated,
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
