package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ozontech/seq-ui/internal/app/config/v2"
)

func TestJWTProvider(t *testing.T) {
	expInterval := time.Second
	p1 := NewJWTProvider(&config.JWT{SecretKey: "secret-1"})
	p2 := NewJWTProvider(&config.JWT{SecretKey: "secret-2"})
	// test valid issue token
	tokenName1 := "test-1"
	testStart := time.Now()
	exp1 := testStart.Add(expInterval)
	tokenStr1, err := p1.IssueToken(tokenName1, exp1.Unix())
	require.NoError(t, err)
	require.NotEmpty(t, tokenStr1)
	// test valid verify token
	claims1, err := p1.Verify(tokenStr1)
	require.NoError(t, err)
	require.NotNil(t, claims1)
	require.Equal(t, tokenName1, claims1.Name)
	// test invalid verify, wrong secret
	claims1, err = p2.Verify(tokenStr1)
	require.Error(t, err)
	require.Nil(t, claims1)
	// test invalid verify, token expired
	time.Sleep(expInterval - time.Since(testStart))
	claims1, err = p1.Verify(tokenStr1)
	require.Error(t, err)
	require.Nil(t, claims1)
}
