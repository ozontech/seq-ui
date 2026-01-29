package tracing

import (
	"context"
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	tracerName = "seq-ui"
)

type Config struct {
	ServiceName  string  `env:"TRACING_SERVICE_NAME"`
	AgentHost    string  `env:"TRACING_AGENT_HOST"`
	AgentPort    string  `env:"TRACING_AGENT_PORT"`
	SamplerParam float64 `env:"TRACING_SAMPLER_PARAM"`
}

func Initialize() (*Config, error) {
	tracingCfg, err := readConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to read tracing config: %w", err)
	}

	if err := validateTracingConfig(tracingCfg); err != nil {
		return nil, err
	}

	tp, err := newTracerProvider(tracingCfg)
	if err != nil {
		return nil, fmt.Errorf("can't create tracer provider: %w", err)
	}

	// Register our TracerProvider as the global so any imported
	// instrumentation in the future will default to using it.
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return tracingCfg, nil
}

func StartSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	return otel.GetTracerProvider().Tracer(tracerName).Start(ctx, name)
}

func newTracerProvider(cfg *Config) (*tracesdk.TracerProvider, error) {
	// Create the Jaeger exporter
	exp, err := jaeger.New(
		jaeger.WithAgentEndpoint(
			jaeger.WithAgentHost(cfg.AgentHost),
			jaeger.WithAgentPort(cfg.AgentPort),
		),
	)
	if err != nil {
		return nil, err
	}

	tp := tracesdk.NewTracerProvider(
		tracesdk.WithSampler(tracesdk.TraceIDRatioBased(cfg.SamplerParam)),
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exp),
		// Record information about this application in a Resource.
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(cfg.ServiceName),
		)),
	)

	return tp, nil
}

func TimestampToStringValue(t *timestamppb.Timestamp) attribute.Value {
	if t == nil {
		return attribute.StringValue("nil")
	}
	return attribute.StringValue(t.AsTime().Format(time.DateTime))
}

func validateTracingConfig(cfg *Config) error {
	if cfg.ServiceName == "" {
		return fmt.Errorf("tracing_service_name not found")
	}
	if cfg.AgentHost == "" {
		return fmt.Errorf("tracing_agent_host not found")
	}
	if cfg.AgentPort == "" {
		return fmt.Errorf("tracing_agent_port not found")
	}
	if cfg.SamplerParam < 0 || cfg.SamplerParam > 1 {
		return fmt.Errorf("tracing_sampler_param must be between 0 and 1, got: %f", cfg.SamplerParam)
	}
	return nil
}

func readConfig() (*Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return &Config{}, fmt.Errorf("failed to parse tracing config from environment: %w", err)
	}
	return &cfg, nil
}
