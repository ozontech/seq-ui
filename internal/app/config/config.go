package config

import (
	"bytes"
	"fmt"
	"os"

	"github.com/ozontech/seq-ui/internal/app/config/migrate"
	v1 "github.com/ozontech/seq-ui/internal/app/config/v1"
	v2 "github.com/ozontech/seq-ui/internal/app/config/v2"
	"gopkg.in/yaml.v3"
)

const (
	configV1Name    = "config_v1"
	configV2Name    = "config_v2"
	previousVersion = 1
	currentVersion  = 2
)

type configMeta struct {
	Version *int `yaml:"version"`
}

// FromFile parse config from config path.
func FromFile(cfgPath string) (v2.Config, error) {
	cfgBytes, err := os.ReadFile(cfgPath) //nolint:gosec
	if err != nil {
		return v2.Config{}, fmt.Errorf("error reading file: %s", err)
	}

	meta, err := parse[configMeta](cfgBytes, false)
	if err != nil {
		return v2.Config{}, err
	}

	cfg := v2.Config{}
	switch {
	case meta.Version == nil || *meta.Version == previousVersion:
		cfgV1, err := parse[v1.Config](cfgBytes, true)
		if err != nil {
			return v2.Config{}, err
		}
		cfg = migrate.V1ToV2(cfgV1)
	case *meta.Version == currentVersion:
		cfg, err = parse[v2.Config](cfgBytes, true)
		if err != nil {
			return v2.Config{}, err
		}
	default:
		return v2.Config{}, fmt.Errorf("unsupported config version: %d", *meta.Version)
	}

	if err := v2.Normalize(&cfg); err != nil {
		return v2.Config{}, fmt.Errorf("normalize config: %w")
	}

	return cfg, nil
}

func parse[T any](cfg []byte, strict bool) (T, error) {
	var result T

	decoder := yaml.NewDecoder(bytes.NewReader(cfg))
	if strict {
		decoder.KnownFields(true)
	}
	if err := decoder.Decode(&result); err != nil {
		return result, fmt.Errorf("error parsing config: %w", err)
	}

	return result, nil
}
