package tracing

import (
	"context"
	"fmt"
	"time"

	"github.com/ozontech/seq-ui/internal/app/config"
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

func Initialize(cfg *config.Tracing) error {
	if cfg == nil {
		return nil
	}

	if err := validateTracingConfig(cfg); err != nil {
		return err
	}

	tp, err := newTracerProvider(cfg)
	if err != nil {
		return fmt.Errorf("can't create tracer provider: %w", err)
	}

	// Register our TracerProvider as the global so any imported
	// instrumentation in the future will default to using it.
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return nil
}

func StartSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	return otel.GetTracerProvider().Tracer(tracerName).Start(ctx, name)
}

func newTracerProvider(cfg *config.Tracing) (*tracesdk.TracerProvider, error) {
	// Create the Jaeger exporter
	exp, err := jaeger.New(
		jaeger.WithAgentEndpoint(
			jaeger.WithAgentHost(cfg.Agent.Host),
			jaeger.WithAgentPort(cfg.Agent.Port),
		),
	)
	if err != nil {
		return nil, err
	}

	tp := tracesdk.NewTracerProvider(
		tracesdk.WithSampler(tracesdk.TraceIDRatioBased(cfg.Sampler.Param)),
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exp),
		// Record information about this application in a Resource.
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(cfg.Resource.ServiceName),
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

func validateTracingConfig(cfg *config.Tracing) error {
	if cfg.Resource.ServiceName == "" {
		return fmt.Errorf("tracing resource.service_name not found")
	}
	if cfg.Agent.Host == "" {
		return fmt.Errorf("tracing agent.host not found")
	}
	if cfg.Agent.Port == "" {
		return fmt.Errorf("tracing agent.port not found")
	}
	if cfg.Sampler.Param < 0 || cfg.Sampler.Param > 1 {
		return fmt.Errorf("tracing sampler.param must be between 0 and 1, got: %f", cfg.Sampler.Param)
	}
	return nil
}
