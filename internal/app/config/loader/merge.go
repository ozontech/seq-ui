package loader

import (
	"fmt"
	"maps"
	"os"

	"gopkg.in/yaml.v3"
)

// mergeFiles merges configs in the given order, later files override earlier ones.
func mergeFiles(cfgPaths []string) (map[string]any, int, error) {
	mergedConfig := make(map[string]any, len(cfgPaths))
	version := 0

	for i, path := range cfgPaths {
		yamlContents, err := os.ReadFile(path) //nolint:gosec
		if err != nil {
			return nil, 0, fmt.Errorf("read config file %q: %w", path, err)
		}

		v, err := readVersion(yamlContents)
		if err != nil {
			return nil, 0, fmt.Errorf("read config %q version: %w", path, err)
		}
		if i > 0 && v != version {
			return nil, 0, fmt.Errorf("different config versions: %d and %d", version, v)
		}
		version = v

		var currentCfg map[string]any
		if err := yaml.Unmarshal(yamlContents, &currentCfg); err != nil {
			return nil, 0, fmt.Errorf("parse config file yaml %q: %w", path, err)
		}

		mergedConfig = mergeYAMLs(mergedConfig, currentCfg)
	}

	return mergedConfig, version, nil
}

func mergeEnvOptions(root map[string]any) {
	mergeOptions := func(root map[string]any) {
		rootOpts, _ := root["options"].(map[string]any)
		envs, _ := root["envs"].(map[string]any)
		if rootOpts == nil || len(envs) == 0 {
			return
		}

		for name, e := range envs {
			envOpts, _ := e.(map[string]any)
			envs[name] = mergeYAMLs(rootOpts, envOpts)
		}
	}

	handlers, _ := root["handlers"].(map[string]any)
	for _, h := range handlers {
		if handler, _ := h.(map[string]any); handler != nil {
			mergeOptions(handler)
		}
	}

	server, _ := root["server"].(map[string]any)
	if auth, _ := server["auth"].(map[string]any); auth != nil {
		mergeOptions(auth)
	}
	if rl, _ := server["rate_limiting"].(map[string]any); rl != nil {
		mergeOptions(rl)
	}
}

func mergeYAMLs(base, override map[string]any) map[string]any {
	merged := make(map[string]any)
	maps.Copy(merged, base)

	for k, v := range override {
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
