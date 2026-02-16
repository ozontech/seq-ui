package tracing

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateTracingConfig(t *testing.T) {
	tCases := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name: "valid_tracing_config",
			cfg: &Config{
				ServiceName:  "seq-ui",
				AgentHost:    "localhost",
				AgentPort:    "6831",
				SamplerParam: 0.7,
			},
			wantErr: false,
		},
		{
			name: "missing_service_name",
			cfg: &Config{
				AgentHost:    "localhost",
				AgentPort:    "6831",
				SamplerParam: 0.7,
			},
			wantErr: true,
		},
		{
			name: "missing_agent_host",
			cfg: &Config{
				ServiceName:  "seq-ui",
				AgentPort:    "6831",
				SamplerParam: 0.7,
			},
			wantErr: true,
		},
		{
			name: "missing_agent_port",
			cfg: &Config{
				ServiceName:  "seq-ui",
				AgentHost:    "localhost",
				SamplerParam: 0.7,
			},
			wantErr: true,
		},
		{
			name: "sampler_param_too_low",
			cfg: &Config{
				ServiceName:  "seq-ui",
				AgentHost:    "localhost",
				AgentPort:    "6831",
				SamplerParam: -1.5,
			},
			wantErr: true,
		},
		{
			name: "sampler_param_too_high",
			cfg: &Config{
				ServiceName:  "seq-ui",
				AgentHost:    "localhost",
				AgentPort:    "6831",
				SamplerParam: 1.5,
			},
			wantErr: true,
		},
	}

	for _, tCase := range tCases {
		tCase := tCase
		t.Run(tCase.name, func(t *testing.T) {
			t.Parallel()
			err := validateTracingConfig(tCase.cfg)
			require.Equal(t, tCase.wantErr, err != nil)
		})
	}
}
