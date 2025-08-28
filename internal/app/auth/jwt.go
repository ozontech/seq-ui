package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type JWTClaims struct {
	Name string `json:"name"`

	jwt.StandardClaims
}

type JWTProvider interface {
	Verify(token string) (*JWTClaims, error)
	IssueToken(name string, exp int64) (string, error)
}

type jwtProvider struct {
	secretKey string
}

func NewJWTProvider(secretKey string) JWTProvider {
	return &jwtProvider{
		secretKey: secretKey,
	}
}

// Verify checks whether the token was created using the configuration of this provider and is still valid.
func (j *jwtProvider) Verify(token string) (*JWTClaims, error) {
	parsedToken, err := jwt.ParseWithClaims(token, &JWTClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(j.secretKey), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := parsedToken.Claims.(*JWTClaims); ok && parsedToken.Valid {
		if err := claims.Valid(); err != nil {
			return nil, fmt.Errorf("invalid token: %w", err)
		}
		return claims, nil
	}
	return nil, errors.New("invalid token: invalid claims")
}

// IssueToken maps args to claims and creates a new signed token string.
// Expiration must be passed as unix timestamp in seconds.
func (j *jwtProvider) IssueToken(name string, exp int64) (string, error) {
	claims := jwt.MapClaims{
		"name": name,
		"iat":  time.Now().Unix(),
	}
	if exp > 0 {
		claims["exp"] = exp
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(j.secretKey))
	if err != nil {
		return "", err
	}
	return tokenStr, nil
}
