package seqdb

import (
	"context"
	"errors"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/pkg/client/seqdb/seqproxyapi/v1"
	mock "github.com/ozontech/seq-ui/internal/pkg/client/seqdb/seqproxyapi/v1/mock"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Test_GRPCClient_Export(t *testing.T) {
	from := time.Now()
	to := from.Add(time.Second)
	var limit int32 = 3

	eventTime := time.Date(2024, time.December, 31, 10, 20, 30, 400000, time.UTC) // 2024-12-31T10:20:30.0004Z
	eventTimePB := timestamppb.New(eventTime)

	type mockArgs struct {
		req  *seqproxyapi.ExportRequest
		resp *mock.MockSeqProxyApi_ExportClient
		err  error
	}

	prepareMockArgs := func(ctrl *gomock.Controller, req *seqapi.ExportRequest, docs []seqproxyapi.Document, errType streamErrorType) mockArgs {
		var proxyReq *seqproxyapi.ExportRequest
		var proxyResp *mock.MockSeqProxyApi_ExportClient

		if req != nil {
			proxyReq = &seqproxyapi.ExportRequest{
				Query:  makeProxySearchQuery(req.Query, req.From, req.To),
				Size:   int64(req.Limit),
				Offset: int64(req.Offset),
			}
		}

		if errType == streamErrProxy {
			return mockArgs{
				req:  proxyReq,
				resp: nil,
				err:  errors.New("proxy error"),
			}
		}

		proxyResp = mock.NewMockSeqProxyApi_ExportClient(ctrl)
		if errType == streamErrRecv {
			proxyResp.EXPECT().Recv().Return(nil, errors.New("recv error")).Times(1)
		} else {
			for i := range docs {
				proxyResp.EXPECT().Recv().Return(&seqproxyapi.ExportResponse{
					Doc: &seqproxyapi.Document{
						Id:   docs[i].Id,
						Data: docs[i].Data,
						Time: docs[i].Time,
					},
				}, nil).Times(1)
			}
			proxyResp.EXPECT().Recv().Return(nil, io.EOF).Times(1)
		}

		return mockArgs{
			req:  proxyReq,
			resp: proxyResp,
			err:  nil,
		}
	}

	tests := []struct {
		name string

		req      *seqapi.ExportRequest
		docs     []seqproxyapi.Document
		wantResp string
		wantErr  streamErrorType
	}{
		{
			name: "ok_jsonl",
			req: &seqapi.ExportRequest{
				Query:  "test_ok",
				From:   timestamppb.New(from),
				To:     timestamppb.New(to),
				Limit:  limit,
				Offset: 0,
			},
			docs: []seqproxyapi.Document{
				{Id: "test1", Data: []byte(`{"key1":"val1","key2":"val2"}`), Time: eventTimePB},
				{Id: "test2", Data: []byte(`{"key1":"val10","key2":"val20"}`), Time: eventTimePB},
				{Id: "test3", Data: []byte(`{"key1":"val100","key2":"val200"}`), Time: eventTimePB},
			},
			wantResp: "{\"id\":\"test1\",\"data\":{\"key1\":\"val1\",\"key2\":\"val2\"},\"time\":\"2024-12-31T10:20:30.0004Z\"}\r\n{\"id\":\"test2\",\"data\":{\"key1\":\"val10\",\"key2\":\"val20\"},\"time\":\"2024-12-31T10:20:30.0004Z\"}\r\n{\"id\":\"test3\",\"data\":{\"key1\":\"val100\",\"key2\":\"val200\"},\"time\":\"2024-12-31T10:20:30.0004Z\"}\r\n",
		},
		{
			name: "ok_jsonl_empty",
			req: &seqapi.ExportRequest{
				Query:  "test_ok_empty",
				From:   timestamppb.New(from),
				To:     timestamppb.New(to),
				Limit:  limit,
				Offset: 0,
			},
		},
		{
			name: "ok_csv",
			req: &seqapi.ExportRequest{
				Query:  "test_ok",
				From:   timestamppb.New(from),
				To:     timestamppb.New(to),
				Limit:  limit,
				Offset: 0,
				Format: seqapi.ExportFormat_EXPORT_FORMAT_CSV,
				Fields: []string{"key1", "key3"},
			},
			docs: []seqproxyapi.Document{
				{Id: "test1", Data: []byte(`{"key1":"val1,a","key2":"val2,b","key3":"val3,c"}`), Time: eventTimePB},
				{Id: "test2", Data: []byte(`{"key1":"val10,a","key2":"val20,b","key3":"val30,c"}`), Time: eventTimePB},
				{Id: "test3", Data: []byte(`{"key1":"val100,a","key2":"val200,b","key3":"val300,c"}`), Time: eventTimePB},
			},
			wantResp: "key1,key3\r\n\"val1,a\",\"val3,c\"\r\n\"val10,a\",\"val30,c\"\r\n\"val100,a\",\"val300,c\"\r\n",
		},
		{
			name: "ok_csv_with_skips",
			req: &seqapi.ExportRequest{
				Query:  "test_ok_skips",
				From:   timestamppb.New(from),
				To:     timestamppb.New(to),
				Limit:  limit,
				Offset: 0,
				Format: seqapi.ExportFormat_EXPORT_FORMAT_CSV,
				Fields: []string{"some-key1", "key2"},
			},
			docs: []seqproxyapi.Document{
				{Id: "test1", Data: []byte(`{"key1":"val1,a","key2":"val2,b","key3":"val3,c"}`), Time: eventTimePB},
				{Id: "test2", Data: []byte(`{"key1":"val10,a","key3":"val30,c"}`), Time: eventTimePB},
				{Id: "test3", Data: []byte(`{"key1":"val100,a","key2":"val200,b","key3":"val300,c"}`), Time: eventTimePB},
			},
			wantResp: "some-key1,key2\r\n,\"val2,b\"\r\n,\r\n,\"val200,b\"\r\n",
		},
		{
			name: "ok_csv_escaped_quotes",
			req: &seqapi.ExportRequest{
				Query:  "test_ok_escaped",
				From:   timestamppb.New(from),
				To:     timestamppb.New(to),
				Limit:  limit,
				Offset: 0,
				Format: seqapi.ExportFormat_EXPORT_FORMAT_CSV,
				Fields: []string{"key1", "key3"},
			},
			docs: []seqproxyapi.Document{
				{Id: "test1", Data: []byte(`{"key1":"val1,a","key2":"val2,b","key3":"test \"quoted\""}`), Time: eventTimePB},
			},
			wantResp: "key1,key3\r\n\"val1,a\",\"test \\\"\"quoted\\\"\"\"\r\n",
		},
		{
			name: "err_proxy",
			req: &seqapi.ExportRequest{
				Query:  "test_err_proxy",
				From:   timestamppb.New(from),
				To:     timestamppb.New(to),
				Limit:  limit,
				Offset: 0,
			},
			wantErr: streamErrProxy,
		},
		{
			name: "err_stream_recv",
			req: &seqapi.ExportRequest{
				Query:  "test_err_stream_recv",
				From:   timestamppb.New(from),
				To:     timestamppb.New(to),
				Limit:  limit,
				Offset: 0,
			},
			wantErr: streamErrRecv,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			ctrl := gomock.NewController(t)

			mArgs := prepareMockArgs(ctrl, tt.req, tt.docs, tt.wantErr)

			seqProxyMock := mock.NewMockSeqProxyApiClient(ctrl)
			seqProxyMock.EXPECT().Export(ctx, mArgs.req).
				Return(mArgs.resp, mArgs.err).Times(1)

			c := initGRPCClient(seqProxyMock)

			w := httptest.NewRecorder()
			cw, err := httputil.NewChunkedWriter(w)
			assert.NoError(t, err)

			err = c.Export(ctx, tt.req, cw)

			require.Equal(t, tt.wantErr != streamErrNo, err != nil)
			if tt.wantErr != streamErrNo {
				return
			}

			res := w.Result()
			defer res.Body.Close()

			data, err := io.ReadAll(res.Body)
			assert.NoError(t, err)

			require.Equal(t, tt.wantResp, string(data))
		})
	}
}
