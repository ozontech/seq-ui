package http

import (
	"context"
	"net/http"
)

type envContextKey struct{}

func getEnvFromContext(ctx context.Context) string {
	if v := ctx.Value(envContextKey{}); v != nil {
		return v.(string)
	}
	return ""
}

func (a *API) envInterceptor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		env := r.URL.Query().Get("env")
		ctx := context.WithValue(r.Context(), envContextKey{}, env)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
