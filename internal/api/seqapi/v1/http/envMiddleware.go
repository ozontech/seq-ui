package http

import (
	"context"
	"net/http"
)

// Была такая ошибка от staticcheck:
// A1029: should not use built-in type string as key for value; define your own type to avoid collisions
// Поэтому здесь кастомку создал
type contextKey string

const (
	envContextKey contextKey = "env"
)

func (a *API) envMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/envs" {
			next.ServeHTTP(w, r)
			return
		}

		env := r.URL.Query().Get("env")
		// Валидация между будет
		ctx := context.WithValue(r.Context(), envContextKey, env)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
