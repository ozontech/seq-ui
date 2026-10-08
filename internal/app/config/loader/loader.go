package loader

import (
	"bytes"
	"errors"
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

	version, err := readVersion(cfgBytes)
	if err != nil {
		return v2.Config{}, fmt.Errorf("read version: %w", err)
	}

	switch version {
	case V1:
		cfgV1, err := decode[v1.Config](cfgBytes, true)
		if err != nil {
			return v2.Config{}, fmt.Errorf("parse config v1: %w", err)
		}
		cfgBytes, err = encode(migrate.V1ToV2(cfgV1))
		if err != nil {
			return v2.Config{}, err
		}
	case V2:
	default:
		return v2.Config{}, fmt.Errorf("unsupported config version: %d", version)
	}

	cfgBytes, err = mergeEnvOptions(cfgBytes)
	if err != nil {
		return v2.Config{}, fmt.Errorf("merge env options: %w", err)
	}

	cfg, err := decode[v2.Config](cfgBytes, true)
	if err != nil {
		return v2.Config{}, fmt.Errorf("parse config v2: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return v2.Config{}, fmt.Errorf("normalize config: %w", err)
	}

	return cfg, nil
}

func ToLatestVersion(cfgBytes []byte) ([]byte, error) {
	version, err := readVersion(cfgBytes)
	if err != nil {
		return nil, fmt.Errorf("read version: %w", err)
	}

	switch version {
	case V1:
		cfgV1, err := decode[v1.Config](cfgBytes, true)
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

func readVersion(cfgBytes []byte) (int, error) {
	meta, err := decode[configMeta](cfgBytes, false)
	if err != nil {
		return 0, err
	}

	if meta.Version == nil {
		return V1, nil
	}

	return *meta.Version, nil
}

func decode[T any](cfg []byte, strict bool) (T, error) {
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
