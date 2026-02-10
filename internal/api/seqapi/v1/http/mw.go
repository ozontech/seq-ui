package http

import (
	"context"
	"net/http"
)

type envContextKey struct{}

func setEnvToContext(ctx context.Context, env string) context.Context {
	return context.WithValue(ctx, envContextKey{}, env)
}

//nolint:unused
func getEnvFromContext(ctx context.Context) string {
	if v := ctx.Value(envContextKey{}); v != nil {
		return v.(string)
	}
	return ""
}

func (a *API) envInterceptor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/envs" {
			next.ServeHTTP(w, r)
			return
		}

		env := r.URL.Query().Get("env")
		// Валидация между будет
		ctx := setEnvToContext(r.Context(), env)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
