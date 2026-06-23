package http

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/api/seqapi/v1/test"
	mock_seqdb "github.com/ozontech/seq-ui/internal/pkg/client/seqdb/mock"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
)

func TestServeGetHistogram(t *testing.T) {
	type mockArgs struct {
		req  *seqapi.GetHistogramRequest
		resp *seqapi.GetHistogramResponse
		err  error
	}

	tests := []struct {
		name string

		req     getHistogramRequest
		want    getHistogramResponse
		wantErr bool

		mockArgs *mockArgs
	}{
		{
			name: "ok",
			req: getHistogramRequest{
				Query:    testQuery,
				From:     testTimestamp,
				To:       testTimestamp.Add(time.Second),
				Interval: "5s",
			},
			want: getHistogramResponse{
				Histogram: histogram{
					Buckets: histogramBuckets{
						{Key: "0", DocCount: "1"},
						{Key: "100", DocCount: "2"},
					},
				},
				Error:           apiError{Code: aecNo},
				PartialResponse: false,
			},
			mockArgs: &mockArgs{
				req: &seqapi.GetHistogramRequest{
					Query:    testQuery,
					From:     timestamppb.New(testTimestamp),
					To:       timestamppb.New(testTimestamp.Add(time.Second)),
					Interval: "5s",
				},
				resp: &seqapi.GetHistogramResponse{
					Histogram: test.MakeHistogram(2),
					Error: &seqapi.Error{
						Code: seqapi.ErrorCode_ERROR_CODE_NO,
					},
				},
			},
		},
		{
			name: "err_partial_response",
			req: getHistogramRequest{
				Query:    testQuery,
				From:     testTimestamp,
				To:       testTimestamp.Add(time.Second),
				Interval: "10s",
			},
			want: getHistogramResponse{
				Histogram: histogram{
					Buckets: histogramBuckets{},
				},
				Error:           apiError{Code: aecPartialResponse, Message: "partial response"},
				PartialResponse: true,
			},
			mockArgs: &mockArgs{
				req: &seqapi.GetHistogramRequest{
					Query:    testQuery,
					From:     timestamppb.New(testTimestamp),
					To:       timestamppb.New(testTimestamp.Add(time.Second)),
					Interval: "10s",
				},
				resp: &seqapi.GetHistogramResponse{
					Error: &seqapi.Error{
						Code:    seqapi.ErrorCode_ERROR_CODE_PARTIAL_RESPONSE,
						Message: "partial response",
					},
					PartialResponse: true,
				},
			},
		},
		{
			name: "err_client",
			req: getHistogramRequest{
				Query:    testQuery,
				From:     testTimestamp,
				To:       testTimestamp.Add(time.Second),
				Interval: "20s",
			},
			wantErr: true,
			mockArgs: &mockArgs{
				req: &seqapi.GetHistogramRequest{
					Query:    testQuery,
					From:     timestamppb.New(testTimestamp),
					To:       timestamppb.New(testTimestamp.Add(time.Second)),
					Interval: "20s",
				},
				err: errors.New("client error"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			seqData := test.APITestData{}

			if tt.mockArgs != nil {
				ctrl := gomock.NewController(t)
				seqDbMock := mock_seqdb.NewMockClient(ctrl)

				seqDbMock.EXPECT().
					GetHistogram(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.resp, tt.mockArgs.err).
					Times(1)

				seqData.Mocks.SeqDB = seqDbMock
			}

			api := setupTestAPI(seqData)

			httputil.DoTestHTTPEx(t, httputil.TestDataHTTPEx[getHistogramRequest, getHistogramResponse]{
				Method:  http.MethodPost,
				Target:  "/seqapi/v1/histogram",
				Req:     tt.req,
				Handler: api.serveGetHistogram,
				Want:    tt.want,
				WantErr: tt.wantErr,
			})
		})
	}
}
