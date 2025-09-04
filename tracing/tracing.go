package tracing

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

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

var serviceName string

type Settings struct {
	ServiceName  string
	JaegerAgent  JaegerAgent
	SamplerParam float64
}

type JaegerAgent struct {
	Host string
	Port string
}

func Initialize() (Settings, error) {
	tp, settings, err := newTracerProvider()
	if err != nil {
		return Settings{}, fmt.Errorf("can't create trace provider: %w", err)
	}

	// Register our TracerProvider as the global so any imported
	// instrumentation in the future will default to using it.
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return settings, nil
}

func StartSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	return otel.GetTracerProvider().Tracer(serviceName).Start(ctx, name)
}

func newTracerProvider() (*tracesdk.TracerProvider, Settings, error) {
	host, err := getEnvVar("JAEGER_AGENT_HOST")
	if err != nil {
		return nil, Settings{}, err
	}

	port, err := getEnvVar("JAEGER_AGENT_PORT")
	if err != nil {
		return nil, Settings{}, err
	}

	serviceName, err = getEnvVar("TRACING_SERVICE_NAME")
	if err != nil {
		return nil, Settings{}, err
	}

	samplerParamStr, err := getEnvVar("TRACING_SAMPLER_PARAM")
	if err != nil {
		return nil, Settings{}, err
	}

	samplerParam, err := strconv.ParseFloat(samplerParamStr, 64)
	if err != nil {
		return nil, Settings{}, fmt.Errorf("can't parse sampling param: %w", err)
	}

	// Create the Jaeger exporter
	exp, err := jaeger.New(
		jaeger.WithAgentEndpoint(
			jaeger.WithAgentHost(host),
			jaeger.WithAgentPort(port),
		),
	)
	if err != nil {
		return nil, Settings{}, err
	}

	tp := tracesdk.NewTracerProvider(
		tracesdk.WithSampler(tracesdk.TraceIDRatioBased(samplerParam)),
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exp),
		// Record information about this application in a Resource.
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
		)),
	)

	return tp, Settings{
		ServiceName:  serviceName,
		JaegerAgent:  JaegerAgent{Host: host, Port: port},
		SamplerParam: samplerParam,
	}, nil
}

func getEnvVar(name string) (string, error) {
	value, found := os.LookupEnv(name)
	if !found {
		return "", fmt.Errorf("env var '%s' not found", name)
	}
	return value, nil
}

func TimestampToStringValue(t *timestamppb.Timestamp) attribute.Value {
	if t == nil {
		return attribute.StringValue("nil")
	}
	return attribute.StringValue(t.AsTime().Format(time.DateTime))
}
