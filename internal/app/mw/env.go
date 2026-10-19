package mw

import (
	"fmt"

	"github.com/ozontech/seq-ui/internal/app/config/v2"
)

// HandlerEnvs describes envs of the handler.
type handlerEnvs struct {
	DefaultEnv string
	Envs       map[string]struct{}
}

// EnvResolver resolves env of the request for auth and rate limiting by api name.
// APIs not present in the resolver (userprofile, dashboards) or handlers without envs
// always use root options (empty env), env from the request is ignored for them.
type EnvResolver map[string]handlerEnvs

// newEnvResolver collects envs of handlers that support envs.
func NewEnvResolver(handlers *config.Handlers) EnvResolver {
	r := EnvResolver{}

	seqAPIEnvs := handlerEnvs{DefaultEnv: handlers.SeqAPI.DefaultEnv, Envs: envNames(handlers.SeqAPI.Envs)}
	r["seqapi"], r["SeqAPIService"] = seqAPIEnvs, seqAPIEnvs

	if h := handlers.ErrorGroups; h != nil {
		envs := handlerEnvs{DefaultEnv: h.DefaultEnv, Envs: envNames(h.Envs)}
		r["errorgroups"], r["ErrorGroupsService"] = envs, envs
	}
	if h := handlers.MassExport; h != nil {
		envs := handlerEnvs{DefaultEnv: h.DefaultEnv, Envs: envNames(h.Envs)}
		r["massexport"], r["MassExportService"] = envs, envs
	}
	if h := handlers.Admin; h != nil {
		envs := handlerEnvs{DefaultEnv: h.DefaultEnv, Envs: envNames(h.Envs)}
		r["admin"], r["AdminService"] = envs, envs
	}
	if h := handlers.AsyncSearch; h != nil {
		r["async_search"] = handlerEnvs{DefaultEnv: h.DefaultEnv, Envs: envNames(h.Envs)}
	}
	return r
}

func (resolver EnvResolver) resolve(api, env string) (string, error) {
	handler, ok := resolver[api]
	if !ok || len(handler.Envs) == 0 {
		return "", nil
	}

	if env == "" {
		return handler.DefaultEnv, nil
	}

	if _, ok := handler.Envs[env]; !ok {
		return "", fmt.Errorf("env %q not found", env)
	}

	return env, nil
}

func envNames[T any](envs map[string]T) map[string]struct{} {
	names := make(map[string]struct{}, len(envs))
	for name := range envs {
		names[name] = struct{}{}
	}
	return names
}
