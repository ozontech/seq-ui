package mw

import (
	"context"
	"strings"
	"time"

	"github.com/ozontech/seq-ui/tracing"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var nonLogHeaders = map[string]struct{}{"authorization": {}}

// arrStr wrapper for []string logging.
type arrStr []string

// MarshalLogArray called when arrStr object is passed to zap.Array.
func (o arrStr) MarshalLogArray(enc zapcore.ArrayEncoder) error {
	for _, item := range o {
		enc.AppendString(item)
	}
	return nil
}

// mapStrArrStr wrapper for map[string][]string logging.
type mapStrArrStr map[string]arrStr

// MarshalLogObject called when mapStrArrStr object is passed to zap.Object.
func (o mapStrArrStr) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	for key, val := range o {
		_ = enc.AddArray(key, val)
	}
	return nil
}

// prepareHeaderMapForLog prepares header map for log. Returns object of
// mapStrArrStr type which has methods for conversion to json object.
func prepareHeaderMapForLog(header map[string][]string) mapStrArrStr {
	hMap := mapStrArrStr{}
	for key, val := range header {
		lkey := strings.ToLower(key)
		if _, has := nonLogHeaders[lkey]; has {
			continue
		}
		hMap[key] = val
	}
	return hMap
}

// requestLogArgs unification for gRPC and http log arguments.
type requestLogArgs struct {
	component   string
	header      map[string][]string
	fullMethod  string
	statusCode  string
	took        time.Duration
	requestBody string
	clientError string
	serverError string
	user        string
	details     []interface{}
}

// logRequestBeforeHandler logs incoming request before the handler call.
// Logs info about component, headers, full method, request body and user if provided in context.
func logRequestBeforeHandler(ctx context.Context, logger *tracing.Logger, args requestLogArgs) {
	processedHeader := prepareHeaderMapForLog(args.header)
	logArgs := []zap.Field{
		zap.String("component", args.component),
		zap.Object("header", processedHeader),
		zap.String("full_method", args.fullMethod),
		zap.String("body", args.requestBody),
	}
	if args.user != "" {
		logArgs = append(logArgs, zap.String("user", args.user))
	}
	logger.Info(ctx, "incoming request", logArgs...)
}

// logRequestAfterHandler logs incoming requests processing results after handler call such as
// status code, took, status detail. Also includes info about component, full method and user
// if provided in context.
//
// Can be linked with incoming request log written before handler call via trace_id or span_id.
//
// Depending on the status code the log level is set to:
//
// * Info - no errors
//
// * Warning - client errors (status code 4xx)
//
// * Error - server errors (status code 5xx).
func logRequestAfterHandler(ctx context.Context, logger *tracing.Logger, args requestLogArgs) {
	logArgs := []zap.Field{
		zap.String("component", args.component),
		zap.String("full_method", args.fullMethod),
		zap.String("status_code", args.statusCode),
		zap.String("took", args.took.String()),
	}
	if args.user != "" {
		logArgs = append(logArgs, zap.String("user", args.user))
	}
	if len(args.details) > 0 {
		logArgs = append(logArgs, zap.Any("status_details", args.details))
	}

	switch {
	case args.clientError != "":
		logArgs = append(logArgs, zap.String("client_error", args.clientError))
		logger.Warn(ctx, "client error occurred", logArgs...)
	case args.serverError != "":
		logArgs = append(logArgs, zap.String("server_error", args.serverError))
		logger.Error(ctx, "server error occurred", logArgs...)
	default:
		logger.Info(ctx, "successful request", logArgs...)
	}
}
