package mw

import (
	"context"
	"fmt"
	"time"

	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/logger"
	"github.com/ozontech/seq-ui/metric"
	"github.com/ozontech/seq-ui/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func GRPCLogInterceptor(logger *tracing.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context, req any, info *grpc.UnaryServerInfo,
		h grpc.UnaryHandler,
	) (any, error) {
		md, _ := metadata.FromIncomingContext(ctx)
		rBody, err := protojson.Marshal(req.(proto.Message))
		if err != nil {
			logger.Error(ctx, "failed to marshal request message", zap.Error(err))
		}
		reqLogArgs := requestLogArgs{
			component:   "gRPC",
			header:      md,
			fullMethod:  info.FullMethod,
			requestBody: string(rBody),
		}
		logRequestBeforeHandler(ctx, logger, reqLogArgs)

		start := time.Now()
		resp, err := h(ctx, req)
		took := time.Since(start)

		st, ok := status.FromError(err)
		if !ok {
			st = status.New(codes.Unknown, err.Error())
		}

		reqLogArgs.statusCode = st.Code().String()
		reqLogArgs.took = took

		errType := gRPCRespErrorTypeFromStatusCode(st.Code())

		if errType == respClientError {
			reqLogArgs.clientError = st.Err().Error()
		}
		if errType == respServerError {
			reqLogArgs.serverError = st.Err().Error()
		}
		if userName, err := types.GetUserKey(ctx); err == nil {
			reqLogArgs.user = userName
		}
		if details := st.Details(); len(details) > 0 {
			reqLogArgs.details = details
		}

		logRequestAfterHandler(ctx, logger, reqLogArgs)

		return resp, err
	}
}

func GRPCMetricInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context, req any, info *grpc.UnaryServerInfo,
		h grpc.UnaryHandler,
	) (any, error) {
		svc, method, _ := parseGRPCFullMethod(info.FullMethod)
		parsedMethod := fmt.Sprintf("%s/%s", svc, method)
		metric.ServerRequestReceived.WithLabelValues("grpc", parsedMethod).Inc()
		start := time.Now()
		resp, err := h(ctx, req)
		took := time.Since(start)
		st, ok := status.FromError(err)
		if !ok {
			st = status.New(codes.Unknown, err.Error())
		}
		metric.HandledIncomingRequest(ctx, "grpc", parsedMethod, st.Code().String(), took)
		return resp, err
	}
}

func GRPCTraceInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context, req any, info *grpc.UnaryServerInfo,
		h grpc.UnaryHandler,
	) (any, error) {
		svc, method, _ := parseGRPCFullMethod(info.FullMethod)
		parsedMethod := fmt.Sprintf("%s/%s", svc, method)

		ctx, span := tracing.StartSpan(ctx, parsedMethod)
		defer span.End()

		resp, err := h(ctx, req)
		if err != nil {
			st, _ := status.FromError(err)
			span.SetAttributes(
				attribute.KeyValue{
					Key:   "error",
					Value: attribute.BoolValue(true),
				},
				attribute.KeyValue{
					Key:   "status_code",
					Value: attribute.StringValue(st.Code().String()),
				},
				attribute.KeyValue{
					Key:   "error_message",
					Value: attribute.StringValue(err.Error()),
				},
			)
		}

		return resp, err
	}
}

// nolint: gochecknoglobals
var noAuthGRPCMethods = map[string]map[string]struct{}{
	"SeqAPIService": {
		"GetFields":       {},
		"GetPinnedFields": {},
		"GetLimits":       {},
	},
}

var errUnauth = status.Error(codes.Unauthenticated, types.ErrUnauthenticated.Error())

func GRPCAuthInterceptor(providers *AuthProviders) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context, req any, info *grpc.UnaryServerInfo,
		h grpc.UnaryHandler,
	) (any, error) {
		svc, method, err := parseGRPCFullMethod(info.FullMethod)
		if err != nil {
			logger.Error("failed to parse gRPC FullMethod", zap.Error(err))
			return nil, errUnauth
		}
		if _, noAuth := noAuthGRPCMethods[svc][method]; noAuth {
			return h(ctx, req)
		}
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			logger.Error("no metadata provided")
			return nil, errUnauth
		}
		authHeaderSlice, ok := md["authorization"]
		if !ok || len(authHeaderSlice) == 0 {
			logger.Error("no authorization metadata provided")
			return nil, errUnauth
		}
		_, checkAPIToken := tokenAuthServices[svc]
		userName, err := providers.auth(ctx, authHeaderSlice[0], checkAPIToken)
		if err != nil {
			logger.Error("token auth failed", zap.Error(err))
			return nil, errUnauth
		}
		ctx = context.WithValue(ctx, types.UserKey{}, userName)

		return h(ctx, req)
	}
}

func GRPCRecoverInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context, req any, info *grpc.UnaryServerInfo,
		h grpc.UnaryHandler,
	) (_ any, err error) {
		defer func() {
			if r := recover(); r != nil {
				svc, method, _ := parseGRPCFullMethod(info.FullMethod)
				handleRecover(fmt.Sprintf("%s/%s", svc, method), r)
				err = status.Error(codes.Internal, "recover: unexpected server error")
			}
		}()

		return h(ctx, req)
	}
}

func GRPCRateLimitInterceptor(rateLimiters map[string]map[string]RateLimiter) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context, req any, info *grpc.UnaryServerInfo,
		h grpc.UnaryHandler,
	) (any, error) {
		svc, method, err := parseGRPCFullMethod(info.FullMethod)
		if err != nil {
			msg := "failed to parse gRPC FullMethod"
			logger.Error(msg, zap.Error(err))
			return nil, status.Error(codes.Internal, msg)
		}

		// methods without rate limiting = methods without auth
		if _, noRateLimit := noAuthGRPCMethods[svc][method]; noRateLimit {
			return h(ctx, req)
		}

		userToRateLimiter, ok := rateLimiters[svc]
		if !ok {
			return h(ctx, req)
		}

		limited, _, err := handleUserRateLimit(ctx, userToRateLimiter, method)
		if err != nil {
			msg := "failed to rate limit request"
			logger.Error(msg, zap.Error(err))
			return nil, status.Error(codes.Internal, msg)
		}
		if limited {
			logger.Warn("request was rate limited")
			return nil, status.Error(codes.ResourceExhausted, "limit exceeded")
		}

		return h(ctx, req)
	}
}

func GRPCProcessHeadersInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context, req any, _ *grpc.UnaryServerInfo,
		h grpc.UnaryHandler,
	) (any, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return h(ctx, req)
		}

		headerSlice, ok := md[types.UseSeqQLHeader]
		if ok && len(headerSlice) > 0 && headerSlice[0] != "" {
			ctx = context.WithValue(ctx, types.UseSeqQL{}, headerSlice[0])
		}

		return h(ctx, req)
	}
}
