package loader

import (
	"fmt"
	"maps"

	"gopkg.in/yaml.v3"
)

func mergeEnvOptions(cfgBytes []byte) ([]byte, error) {
	var root map[string]any
	if err := yaml.Unmarshal(cfgBytes, &root); err != nil {
		return nil, fmt.Errorf("parse yaml for merge: %w", err)
	}

	merge := func(root map[string]any) {
		rootOpts, _ := root["options"].(map[string]any)
		envs, _ := root["envs"].(map[string]any)
		if rootOpts == nil || len(envs) == 0 {
			return
		}

		for _, e := range envs {
			env, _ := e.(map[string]any)
			if env == nil {
				continue
			}

			envOpts, _ := env["options"].(map[string]any)
			env["options"] = mergeYAMLs(rootOpts, envOpts)
		}
	}

	handlers, _ := root["handlers"].(map[string]any)
	for _, h := range handlers {
		if handler, _ := h.(map[string]any); handler != nil {
			merge(handler)
		}
	}

	server, _ := root["server"].(map[string]any)
	if auth, _ := server["auth"].(map[string]any); auth != nil {
		merge(auth)
	}
	if rl, _ := server["rate_limiters"].(map[string]any); rl != nil {
		merge(rl)
	}

	out, err := yaml.Marshal(root)
	if err != nil {
		return nil, fmt.Errorf("marshal merged yaml: %w", err)
	}
	return out, nil
}

func mergeYAMLs(rootOpts, envOpts map[string]any) map[string]any {
	merged := make(map[string]any)
	maps.Copy(merged, rootOpts)

	for k, v := range envOpts {
		if existingValue, exists := merged[k]; exists {
			if existingMap, ok := existingValue.(map[string]any); ok {
				if newMap, ok := v.(map[string]any); ok {
					merged[k] = mergeYAMLs(existingMap, newMap)
					continue
				}
			}
		}
		merged[k] = v
	}
	return merged
}
