package logger

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// nolint:gochecknoglobals
var Instance *zap.Logger

const defaultLevel = zap.InfoLevel

// nolint:gochecknoinits
func init() {
	var level zapcore.Level
	switch strings.ToLower(os.Getenv("LOG_LEVEL")) {
	case "debug":
		level = zap.DebugLevel
	case "info":
		level = zap.InfoLevel
	case "warn":
		level = zap.WarnLevel
	case "error":
		level = zap.ErrorLevel
	case "fatal":
		level = zap.FatalLevel
	default:
		level = defaultLevel
	}

	atomicLevel := zap.NewAtomicLevelAt(level)

	Instance = zap.New(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(zapcore.EncoderConfig{
				TimeKey:        "ts",
				LevelKey:       "level",
				NameKey:        "logger",
				CallerKey:      "caller",
				MessageKey:     "message",
				StacktraceKey:  "stacktrace",
				LineEnding:     zapcore.DefaultLineEnding,
				EncodeLevel:    zapcore.LowercaseLevelEncoder,
				EncodeTime:     zapcore.ISO8601TimeEncoder,
				EncodeDuration: zapcore.SecondsDurationEncoder,
				EncodeCaller:   zapcore.ShortCallerEncoder,
			}),
			zapcore.AddSync(os.Stdout),
			atomicLevel,
		),
		zap.AddCaller(),
		zap.AddStacktrace(zap.FatalLevel),
	)

	Instance.Info("logger created", zap.String("log_level", level.String()))
}

func Debug(msg string, fields ...zap.Field) {
	Instance.Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	Instance.Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	Instance.Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	Instance.Error(msg, fields...)
}

func Panic(msg string, fields ...zap.Field) {
	Instance.Panic(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	Instance.Fatal(msg, fields...)
}
