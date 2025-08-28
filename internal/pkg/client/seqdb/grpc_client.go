package seqdb

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"path"
	"time"

	grpc_mw "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/internal/pkg/client/seqdb/seqproxyapi/v1"
	"github.com/ozontech/seq-ui/logger"
	"github.com/ozontech/seq-ui/metric"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type grpcSearchResp interface {
	GetError() *seqproxyapi.Error
}

func processClientRequestUnaryInterceptor(
	ctx context.Context,
	method string,
	req interface{},
	reply interface{},
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {
	basePath := path.Base(method)
	metric.SeqDBClientRequestSent.WithLabelValues(basePath).Inc()
	start := time.Now()
	err := invoker(ctx, method, req, reply, cc, opts...)
	took := time.Since(start)
	st, _ := status.FromError(err)
	statusCodeStr := st.Code().String()
	metric.SeqDBClientResponseReceived.WithLabelValues(basePath, statusCodeStr).Inc()
	metric.SeqDBClientRequestDuration.WithLabelValues(basePath, statusCodeStr).Observe(took.Seconds())
	if sResp, ok := reply.(grpcSearchResp); ok {
		errorCode := sResp.GetError().GetCode()
		if errorCode != seqproxyapi.ErrorCode_ERROR_CODE_NO {
			metric.SeqDBClientPartialResponse.WithLabelValues(basePath, errorCode.String()).Inc()
		}
	}
	return err
}

func timeoutUnaryInterceptor(timeout time.Duration) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		return invoker(timeoutCtx, method, req, reply, cc, opts...)
	}
}

func passMetadataUnaryInterceptor(
	ctx context.Context,
	method string,
	req any,
	reply any,
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {
	ctx = passMetadata(ctx)
	return invoker(ctx, method, req, reply, cc, opts...)
}

func processClientRequestStreamInterceptor(
	ctx context.Context,
	desc *grpc.StreamDesc,
	cc *grpc.ClientConn,
	method string,
	streamer grpc.Streamer,
	opts ...grpc.CallOption,
) (grpc.ClientStream, error) {
	metric.SeqDBClientRequestSent.WithLabelValues(path.Base(method)).Inc()
	return streamer(ctx, desc, cc, method, opts...)
}

func passMetadataStreamInterceptor(
	ctx context.Context,
	desc *grpc.StreamDesc,
	cc *grpc.ClientConn,
	method string,
	streamer grpc.Streamer,
	opts ...grpc.CallOption,
) (grpc.ClientStream, error) {
	ctx = passMetadata(ctx)
	return streamer(ctx, desc, cc, method, opts...)
}

func passMetadata(ctx context.Context) context.Context {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		ctx = metadata.NewOutgoingContext(ctx, md)
	} else {
		ctx = metadata.NewOutgoingContext(ctx, metadata.New(nil))
	}
	if useSeqQL := types.GetUseSeqQL(ctx); useSeqQL != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, types.UseSeqQLHeader, useSeqQL)
	}
	return ctx
}

type GRPCClient struct {
	clients             []seqproxyapi.SeqProxyApiClient
	timeout             time.Duration
	initialRetryBackoff time.Duration
	maxRetryBackoff     time.Duration
	reqRetries          int
}

func NewGRPCClient(ctx context.Context, params ClientParams) (*GRPCClient, error) {
	if len(params.Addrs) == 0 {
		panic("addrs is empty")
	}

	clients := make([]seqproxyapi.SeqProxyApiClient, 0, len(params.Addrs))
	unaryInterceptors := []grpc.UnaryClientInterceptor{
		processClientRequestUnaryInterceptor,
		passMetadataUnaryInterceptor,
		timeoutUnaryInterceptor(params.Timeout),
	}
	streamInterceptors := []grpc.StreamClientInterceptor{
		processClientRequestStreamInterceptor,
		passMetadataStreamInterceptor,
	}
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(params.MaxRecvMsgSize),
		),
		grpc.WithUnaryInterceptor(grpc_mw.ChainUnaryClient(unaryInterceptors...)),
		grpc.WithStreamInterceptor(grpc_mw.ChainStreamClient(streamInterceptors...)),
	}
	if params.GRPCKeepaliveParams != nil {
		kp := keepalive.ClientParameters{
			Time:                params.GRPCKeepaliveParams.Time,
			Timeout:             params.GRPCKeepaliveParams.Timeout,
			PermitWithoutStream: params.GRPCKeepaliveParams.PermitWithoutStream,
		}
		opts = append(opts,
			grpc.WithKeepaliveParams(kp),
		)
	}
	for _, addr := range params.Addrs {
		conn, err := grpc.DialContext(ctx, addr, opts...)
		if err != nil {
			return nil, fmt.Errorf("failed to dial grpc server addr=%s", addr)
		}
		client := seqproxyapi.NewSeqProxyApiClient(conn)
		clients = append(clients, client)
	}

	reqRetries := params.MaxRetries
	if reqRetries < 0 {
		logger.Warn("setting requests max retries to 0", zap.Int("max_retries", params.MaxRetries))
		reqRetries = 0
	}
	initialRetryBackoff := params.InitialRetryBackoff
	if initialRetryBackoff < 0 {
		logger.Warn("setting requests initial retry backoff to 0", zap.Duration("initial_retry_backoff", params.InitialRetryBackoff))
		initialRetryBackoff = 0
	}
	maxRetryBackoff := params.MaxRetryBackoff
	if maxRetryBackoff < 0 {
		logger.Warn("setting requests max retry backoff to 0", zap.Duration("max_retry_backoff", params.MaxRetryBackoff))
		maxRetryBackoff = 0
	}

	return &GRPCClient{
		clients:             clients,
		timeout:             params.Timeout,
		reqRetries:          reqRetries,
		initialRetryBackoff: initialRetryBackoff,
		maxRetryBackoff:     maxRetryBackoff,
	}, nil
}

type grpcReqFn func(seqproxyapi.SeqProxyApiClient) (any, error)

func (c *GRPCClient) sendRequest(ctx context.Context, reqFn grpcReqFn) (any, error) {
	clients := make([]seqproxyapi.SeqProxyApiClient, len(c.clients))
	copy(clients, c.clients)
	tryFn := func() (any, bool, error) {
		var err error
		var resp any
		rand.Shuffle(len(clients), func(i, j int) {
			clients[i], clients[j] = clients[j], clients[i]
		})

		for _, client := range clients {
			select {
			case <-ctx.Done():
				if errors.Is(ctx.Err(), context.DeadlineExceeded) {
					return nil, false, status.Error(codes.DeadlineExceeded, context.DeadlineExceeded.Error())
				}
				return nil, false, status.Error(codes.Canceled, context.Canceled.Error())
			default:
			}
			resp, err = reqFn(client)
			if err == nil {
				return resp, false, nil
			}
			st, _ := status.FromError(err)
			if st.Code() != codes.Unavailable {
				return resp, false, err
			}
		}

		return resp, true, err
	}

	grpcResp, ok, err := trySendRequestWithBackoff(ctx, tryFn,
		tryWithBackoffParams{
			maxRetries:          c.reqRetries,
			initialRetryBackoff: c.initialRetryBackoff,
			maxRetryBackoff:     c.maxRetryBackoff,
		},
	)

	if !ok {
		return nil, status.Errorf(codes.Internal, "grpc client send request: %v", err)
	}
	return grpcResp, err
}
