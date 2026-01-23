package tracing

import (
	"testing"

	"github.com/ozontech/seq-ui/internal/app/config"
	"github.com/stretchr/testify/require"
)

func TestValidateTracingConfig(t *testing.T) {
	tCases := []struct {
		name    string
		cfg     *config.Tracing
		wantErr bool
	}{
		{
			name: "valid_config",
			cfg: &config.Tracing{
				Resource: config.TracingResource{
					ServiceName: "seq-ui",
				},
				Agent: config.TracingAgent{
					Host: "localhost",
					Port: "6831",
				},
				Sampler: config.TracingSampler{
					Param: 0.7,
				},
			},
			wantErr: false,
		},
		{
			name: "missing_service_name",
			cfg: &config.Tracing{
				Agent: config.TracingAgent{
					Host: "localhost",
					Port: "6831",
				},
				Sampler: config.TracingSampler{
					Param: 0.7,
				},
			},
			wantErr: true,
		},
		{
			name: "missing_agent_host",
			cfg: &config.Tracing{
				Resource: config.TracingResource{
					ServiceName: "seq-ui",
				},
				Agent: config.TracingAgent{
					Port: "6831",
				},
				Sampler: config.TracingSampler{
					Param: 0.5,
				},
			},
			wantErr: true,
		},
		{
			name: "missing_agent_port",
			cfg: &config.Tracing{
				Resource: config.TracingResource{
					ServiceName: "seq-ui",
				},
				Agent: config.TracingAgent{
					Host: "localhost",
				},
				Sampler: config.TracingSampler{
					Param: 0.5,
				},
			},
			wantErr: true,
		},
		{
			name: "sampler_param_too_low",
			cfg: &config.Tracing{
				Resource: config.TracingResource{
					ServiceName: "seq-ui",
				},
				Agent: config.TracingAgent{
					Host: "localhost",
					Port: "6831",
				},
				Sampler: config.TracingSampler{
					Param: -1.5,
				},
			},
			wantErr: true,
		},
		{
			name: "sampler_param_too_high",
			cfg: &config.Tracing{
				Resource: config.TracingResource{
					ServiceName: "seq-ui",
				},
				Agent: config.TracingAgent{
					Host: "localhost",
					Port: "6831",
				},
				Sampler: config.TracingSampler{
					Param: 1.5,
				},
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
