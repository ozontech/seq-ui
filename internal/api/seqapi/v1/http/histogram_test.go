package http

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/api/seqapi/v1/test"
	mock_seqdb "github.com/ozontech/seq-ui/internal/pkg/client/seqdb/mock"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestServeGetHistogram(t *testing.T) {
	query := "message:error"
	from := time.Date(2023, time.September, 25, 10, 20, 30, 0, time.UTC)
	to := from.Add(time.Second)

	formatReqBody := func(interval string) string {
		return fmt.Sprintf(`{"query":%q,"from":%q,"to":%q,"interval":%q}`,
			query, from.Format(time.RFC3339), to.Format(time.RFC3339), interval)
	}

	type mockArgs struct {
		req  *seqapi.GetHistogramRequest
		resp *seqapi.GetHistogramResponse
		err  error
	}

	tests := []struct {
		name string

		reqBody      string
		wantRespBody string
		wantStatus   int

		mockArgs *mockArgs
	}{
		{
			name:    "ok",
			reqBody: formatReqBody("5s"),
			mockArgs: &mockArgs{
				req: &seqapi.GetHistogramRequest{
					Query:    query,
					From:     timestamppb.New(from),
					To:       timestamppb.New(to),
					Interval: "5s",
				},
				resp: &seqapi.GetHistogramResponse{
					Histogram: test.MakeHistogram(2),
					Error: &seqapi.Error{
						Code: seqapi.ErrorCode_ERROR_CODE_NO,
					},
				},
			},
			wantRespBody: `{"histogram":{"buckets":[{"key":"0","docCount":"1"},{"key":"100","docCount":"2"}]},"error":{"code":"ERROR_CODE_NO"},"partialResponse":false}`,
			wantStatus:   http.StatusOK,
		},
		{
			name:    "err_partial_response",
			reqBody: formatReqBody("10s"),
			mockArgs: &mockArgs{
				req: &seqapi.GetHistogramRequest{
					Query:    query,
					From:     timestamppb.New(from),
					To:       timestamppb.New(to),
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
			wantRespBody: `{"histogram":{"buckets":[]},"error":{"code":"ERROR_CODE_PARTIAL_RESPONSE","message":"partial response"},"partialResponse":true}`,
			wantStatus:   http.StatusOK,
		},
		{
			name:       "err_invalid_request",
			reqBody:    "invalid-request",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "err_client",
			reqBody: formatReqBody("20s"),
			mockArgs: &mockArgs{
				req: &seqapi.GetHistogramRequest{
					Query:    query,
					From:     timestamppb.New(from),
					To:       timestamppb.New(to),
					Interval: "20s",
				},
				err: errors.New("client error"),
			},
			wantStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			seqData := test.APITestData{}

			if tt.mockArgs != nil {
				ctrl := gomock.NewController(t)

				seqDbMock := mock_seqdb.NewMockClient(ctrl)
				seqDbMock.EXPECT().GetHistogram(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.resp, tt.mockArgs.err).Times(1)

				seqData.Mocks.SeqDB = seqDbMock
			}

			api := initTestAPI(seqData)
			req := httptest.NewRequest(http.MethodPost, "/seqapi/v1/histogram", strings.NewReader(tt.reqBody))

			httputil.DoTestHTTP(t, httputil.TestDataHTTP{
				Req:          req,
				Handler:      api.serveGetHistogram,
				WantRespBody: tt.wantRespBody,
				WantStatus:   tt.wantStatus,
			})
		})
	}
}
