package seqdb

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/ozontech/seq-ui/internal/pkg/client/seqdb/seqproxyapi/v1"
	mock "github.com/ozontech/seq-ui/internal/pkg/client/seqdb/seqproxyapi/v1/mock"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Test_GRPCClient_GetEvent(t *testing.T) {
	eventTime := timestamppb.New(time.Date(2024, time.December, 31, 10, 20, 30, 400000, time.UTC))

	type mockArgs struct {
		req  *seqproxyapi.FetchRequest
		resp *mock.MockSeqProxyApi_FetchClient
		err  error
	}

	prepareMockArgs := func(ctrl *gomock.Controller, req *seqapi.GetEventRequest, doc *seqproxyapi.Document, errType streamErrorType) mockArgs {
		var proxyReq *seqproxyapi.FetchRequest
		var proxyResp *mock.MockSeqProxyApi_FetchClient

		if req != nil {
			proxyReq = &seqproxyapi.FetchRequest{
				Ids: []string{req.Id},
			}
		}

		if errType == streamErrProxy {
			return mockArgs{
				req:  proxyReq,
				resp: nil,
				err:  errors.New("proxy error"),
			}
		}

		proxyResp = mock.NewMockSeqProxyApi_FetchClient(ctrl)
		if errType == streamErrRecv {
			proxyResp.EXPECT().Recv().Return(nil, errors.New("recv error")).Times(1)
		} else {
			if doc != nil {
				proxyResp.EXPECT().Recv().Return(doc, nil).Times(1)
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

		req *seqapi.GetEventRequest
		doc *seqproxyapi.Document

		wantResp *seqapi.GetEventResponse
		wantErr  streamErrorType
	}{
		{
			name: "ok",
			req: &seqapi.GetEventRequest{
				Id: "test1",
			},
			doc: &seqproxyapi.Document{
				Id:   "test1",
				Data: []byte(`{"key1":"val1","key2":"val2"}`),
				Time: eventTime,
			},
			wantResp: &seqapi.GetEventResponse{
				Event: &seqapi.Event{
					Id:   "test1",
					Data: map[string]string{"key1": "val1", "key2": "val2"},
					Time: eventTime,
				},
			},
		},
		{
			name: "ok_empty",
			req: &seqapi.GetEventRequest{
				Id: "test2",
			},
			wantResp: &seqapi.GetEventResponse{
				Event: &seqapi.Event{},
			},
		},
		{
			name: "ok_invalid_utf8",
			req: &seqapi.GetEventRequest{
				Id: "test3",
			},
			doc: &seqproxyapi.Document{
				Id:   "test3",
				Data: []byte("{\"key1\":\"val1\",\"key2\":\"\xfdval\xff2\xfe\"}"),
				Time: eventTime,
			},
			wantResp: &seqapi.GetEventResponse{
				Event: &seqapi.Event{
					Id:   "test3",
					Data: map[string]string{"key1": "val1", "key2": "�val�2�"},
					Time: eventTime,
				},
			},
		},
		{
			name: "ok_data_null",
			req: &seqapi.GetEventRequest{
				Id: "test4",
			},
			doc: &seqproxyapi.Document{
				Id:   "test4",
				Data: nil,
				Time: eventTime,
			},
			wantResp: &seqapi.GetEventResponse{
				Event: &seqapi.Event{
					Id:   "test4",
					Data: nil,
					Time: eventTime,
				},
			},
		},
		{
			name: "err_proxy",
			req: &seqapi.GetEventRequest{
				Id: "test5",
			},
			wantErr: streamErrProxy,
		},
		{
			name: "err_stream_recv",
			req: &seqapi.GetEventRequest{
				Id: "test6",
			},
			wantErr: streamErrRecv,
		},
		{
			name: "err_convert",
			req: &seqapi.GetEventRequest{
				Id: "test7",
			},
			doc: &seqproxyapi.Document{
				Id:   "test7",
				Data: []byte("invalid-json"),
				Time: eventTime,
			},
			wantErr: streamErrConvert,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mArgs := prepareMockArgs(ctrl, tt.req, tt.doc, tt.wantErr)

			seqProxyMock := mock.NewMockSeqProxyApiClient(ctrl)
			seqProxyMock.EXPECT().Fetch(gomock.Any(), mArgs.req).
				Return(mArgs.resp, mArgs.err).Times(1)

			c := initGRPCClient(seqProxyMock)

			resp, err := c.GetEvent(context.Background(), tt.req)

			require.Equal(t, tt.wantErr != streamErrNo, err != nil)
			if tt.wantResp == nil {
				require.Nil(t, resp)
			} else {
				require.True(t, checkEventsEqual(tt.wantResp.Event, resp.Event))
			}
		})
	}
}
