package tracing

import (
	"context"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type Logger struct {
	logger *zap.Logger
}

func (l *Logger) ZapLogger() *zap.Logger {
	return l.logger
}

func NewLogger(logger *zap.Logger) *Logger {
	return &Logger{
		logger: logger,
	}
}

func (l *Logger) Debug(ctx context.Context, msg string, fields ...zap.Field) {
	l.logger.Debug(msg, getFields(ctx, fields)...)
}

func (l *Logger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	l.logger.Info(msg, getFields(ctx, fields)...)
}

func (l *Logger) Warn(ctx context.Context, msg string, fields ...zap.Field) {
	l.logger.Warn(msg, getFields(ctx, fields)...)
}

func (l *Logger) Error(ctx context.Context, msg string, fields ...zap.Field) {
	l.logger.Error(msg, getFields(ctx, fields)...)
}

func (l *Logger) Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	l.logger.Fatal(msg, getFields(ctx, fields)...)
}

func getFields(ctx context.Context, fields []zap.Field) []zap.Field {
	spanCtx := trace.SpanContextFromContext(ctx)
	if !spanCtx.IsValid() {
		return fields
	}

	return append(fields,
		zap.String("trace_id", spanCtx.TraceID().String()),
		zap.String("span_id", spanCtx.SpanID().String()),
	)
}
