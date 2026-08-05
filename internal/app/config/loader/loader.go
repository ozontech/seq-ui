package loader

import (
	"bytes"
	"fmt"
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

type configMeta struct {
	Version *int `yaml:"version"`
}

// FromFile parse config from config path.
func FromFile(cfgPath string) (v2.Config, error) {
	cfgBytes, err := os.ReadFile(cfgPath) //nolint:gosec
	if err != nil {
		return v2.Config{}, fmt.Errorf("error reading file: %w", err)
	}

	version, err := ReadVersion(cfgBytes)
	if err != nil {
		return v2.Config{}, fmt.Errorf("error reading version: %w", err)
	}

	cfg := v2.Config{}
	switch version {
	case V1:
		cfgV1, err := parse[v1.Config](cfgBytes, true)
		if err != nil {
			return v2.Config{}, fmt.Errorf("error parsing config v1: %w", err)
		}
		cfg = migrate.V1ToV2(cfgV1)
	case V2:
		cfg, err = parse[v2.Config](cfgBytes, true)
		if err != nil {
			return v2.Config{}, fmt.Errorf("error parsing config v2: %w", err)
		}
	default:
		return v2.Config{}, fmt.Errorf("unsupported config version: %d", version)
	}

	if err := v2.Normalize(&cfg); err != nil {
		return v2.Config{}, fmt.Errorf("normalize config: %w", err)
	}

	return cfg, nil
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
