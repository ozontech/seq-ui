package tokenlimiter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLimiter(t *testing.T) {
	t.Parallel()

	user := "unnamed"
	maxTokens := 2

	l := New(maxTokens)

	// check nil channel
	l.Fill(user)

	for range maxTokens {
		require.Equal(t, false, l.Limited(user))
	}
	require.Equal(t, true, l.Limited(user))

	l.Fill(user)
	require.Equal(t, false, l.Limited(user))

	for range maxTokens {
		l.Fill(user)
	}

	require.Equal(t, maxTokens, len(l.m[user]))

	// check filled channel
	l.Fill(user)
}
