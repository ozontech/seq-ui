package loader

import (
	"bytes"
	"errors"
	"fmt"
	"maps"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/ozontech/seq-ui/internal/app/config/migrate"
	v1 "github.com/ozontech/seq-ui/internal/app/config/v1"
	v2 "github.com/ozontech/seq-ui/internal/app/config/v2"
)

const (
	V1 = 1
	V2 = 2
)

var ErrAlreadyLatestConfigVersion = errors.New("config is already at the latest schema version")

type configMeta struct {
	Version *int `yaml:"version"`
}

// FromFile parse config from config path.
func FromFile(cfgPath string) (v2.Config, error) {
	cfgBytes, err := os.ReadFile(cfgPath) //nolint:gosec
	if err != nil {
		return v2.Config{}, fmt.Errorf("read file: %w", err)
	}

	version, err := ReadVersion(cfgBytes)
	if err != nil {
		return v2.Config{}, fmt.Errorf("read version: %w", err)
	}

	switch version {
	case V1:
		cfgV1, err := parse[v1.Config](cfgBytes, true)
		if err != nil {
			return v2.Config{}, fmt.Errorf("parse config v1: %w", err)
		}
		cfgBytes, err = encode(migrate.V1ToV2(cfgV1))
		if err != nil {
			return v2.Config{}, nil
		}
	case V2:
	default:
		return v2.Config{}, fmt.Errorf("unsupported config version: %d", version)
	}

	cfgBytes, err = mergeHandlersEnvOptions(cfgBytes)
	if err != nil {
		return v2.Config{}, fmt.Errorf("merge env options: %w", err)
	}

	cfg, err := parse[v2.Config](cfgBytes, true)
	if err != nil {
		return v2.Config{}, fmt.Errorf("error parsing config v2: %w", err)
	}

	if err := v2.Normalize(&cfg); err != nil {
		return v2.Config{}, fmt.Errorf("normalize config: %w", err)
	}

	return cfg, nil
}

func ToLatestVersion(cfgBytes []byte) ([]byte, error) {
	version, err := ReadVersion(cfgBytes)
	if err != nil {
		return nil, fmt.Errorf("read version: %w", err)
	}

	switch version {
	case V1:
		cfgV1, err := parse[v1.Config](cfgBytes, true)
		if err != nil {
			return nil, fmt.Errorf("parse config v1: %w", err)
		}
		return encode(migrate.V1ToV2(cfgV1))
	case V2:
		return nil, ErrAlreadyLatestConfigVersion
	default:
		return nil, fmt.Errorf("unsupported config version: %d", version)
	}
}

func ReadVersion(cfgBytes []byte) (int, error) {
	meta, err := parse[configMeta](cfgBytes, false)
	if err != nil {
		return 0, err
	}

	if meta.Version == nil {
		return V1, nil
	}

	return *meta.Version, nil
}

func parse[T any](cfg []byte, strict bool) (T, error) {
	var result T

	decoder := yaml.NewDecoder(bytes.NewReader(cfg))
	decoder.KnownFields(strict)
	if err := decoder.Decode(&result); err != nil {
		return result, err
	}

	return result, nil
}

func encode(cfg v2.Config) ([]byte, error) {
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	if err := encoder.Encode(&cfg); err != nil {
		return nil, fmt.Errorf("encode config v2: %w", err)
	}
	if err := encoder.Close(); err != nil {
		return nil, fmt.Errorf("close encoder: %w", err)
	}
	return buf.Bytes(), nil
}

func mergeHandlersEnvOptions(cfgBytes []byte) ([]byte, error) {
	var root map[any]any
	if err := yaml.Unmarshal(cfgBytes, &root); err != nil {
		return nil, fmt.Errorf("parse yaml for merge: %w", err)
	}

	handlers := getMap(root, "handlers")
	for _, h := range handlers {
		handler, _ := h.(map[any]any)
		if handler == nil {
			continue
		}

		rootOpts := getMap(handler, "options")
		envs := getMap(handler, "envs")
		if rootOpts == nil || len(envs) == 0 {
			continue
		}

		for _, e := range envs {
			env, _ := e.(map[any]any)
			if env == nil {
				continue
			}

			envOpts, _ := env["options"].(map[any]any)
			env["options"] = mergeYAMLs(rootOpts, envOpts)
		}
	}

	out, err := yaml.Marshal(root)
	if err != nil {
		return nil, fmt.Errorf("marshal merged yaml: %w", err)
	}
	return out, nil
}

func getMap(root map[any]any, path ...string) map[any]any {
	cur := root
	for _, p := range path {
		next, _ := cur[p].(map[any]any)
		if next == nil {
			return nil
		}
		cur = next
	}

	return cur
}

func mergeYAMLs(a, b map[any]any) map[any]any {
	merged := make(map[any]any)
	maps.Copy(merged, a)

	for k, v := range b {
		if existingValue, exists := merged[k]; exists {
			if existingMap, ok := existingValue.(map[any]any); ok {
				if newMap, ok := v.(map[any]any); ok {
					merged[k] = mergeYAMLs(existingMap, newMap)
					continue
				}
			}
		}
		merged[k] = v
	}
	return merged
}
