package tracing

import (
	"testing"

	"github.com/ozontech/seq-ui/internal/app/config"
)

func TestValidateTracingConfig(t *testing.T) {
	tCases := []struct {
		name    string
		cfg     *config.Tracing
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: &config.Tracing{
				ServiceName: "a-cfg-prov-gw",
				Jaeger: config.TracingJaeger{
					AgentHost: "localhost",
					AgentPort: "6831",
				},
				Sampler: config.TracingSampler{
					Param: 0.7,
				},
			},
			wantErr: false,
		},
		{
			name: "missing service_name",
			cfg: &config.Tracing{
				Jaeger: config.TracingJaeger{
					AgentHost: "localhost",
					AgentPort: "6831",
				},
				Sampler: config.TracingSampler{
					Param: 0.7,
				},
			},
			wantErr: true,
		},
		{
			name: "missing agent_host",
			cfg: &config.Tracing{
				ServiceName: "a-cfg-provider",
				Jaeger: config.TracingJaeger{
					AgentPort: "6831",
				},
				Sampler: config.TracingSampler{
					Param: 0.5,
				},
			},
			wantErr: true,
		},
		{
			name: "missing agent_port",
			cfg: &config.Tracing{
				ServiceName: "ab-admin-gateway",
				Jaeger: config.TracingJaeger{
					AgentHost: "localhost",
				},
				Sampler: config.TracingSampler{
					Param: 0.5,
				},
			},
			wantErr: true,
		},
		{
			name: "sampler param too low",
			cfg: &config.Tracing{
				ServiceName: "ab-controller-public-api",
				Jaeger: config.TracingJaeger{
					AgentHost: "localhost",
					AgentPort: "6831",
				},
				Sampler: config.TracingSampler{
					Param: -1.5,
				},
			},
			wantErr: true,
		},
		{
			name: "sampler param too high",
			cfg: &config.Tracing{
				ServiceName: "ab-events",
				Jaeger: config.TracingJaeger{
					AgentHost: "localhost",
					AgentPort: "6831",
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
			if (err != nil) != tCase.wantErr {
				t.Errorf("validateTracingConfig() error = %v, wantErr %v", err, tCase.wantErr)
			}
		})
	}
}
