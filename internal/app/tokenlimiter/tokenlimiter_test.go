package tokenlimiter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLimiter(t *testing.T) {
	user := "unnamed"
	maxTokens := 2

	l := New(maxTokens)

	// check nil channel
	l.Fill(user)

	for i := 0; i < maxTokens; i++ {
		require.Equal(t, false, l.Limited(user))
	}
	require.Equal(t, true, l.Limited(user))

	l.Fill(user)
	require.Equal(t, false, l.Limited(user))

	for i := 0; i < maxTokens; i++ {
		l.Fill(user)
	}

	require.Equal(t, maxTokens, len(l.m[user]))

	// check filled channel
	l.Fill(user)
}
