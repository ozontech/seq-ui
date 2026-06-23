package http

import (
	"net/http"
	"testing"
	"time"

	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/api/seqapi/v1/test"
	mock_asyncsearches "github.com/ozontech/seq-ui/internal/pkg/service/async_searches/mock"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
)

func TestServeStartAsyncSearch(t *testing.T) {
	var (
		meta = `{"some":"meta"}`
	)

	type mockArgs struct {
		req  *seqapi.StartAsyncSearchRequest
		resp *seqapi.StartAsyncSearchResponse
		err  error
	}

	tests := []struct {
		name string

		req     startAsyncSearchRequest
		want    startAsyncSearchResponse
		wantErr bool

		mockArgs *mockArgs
	}{
		{
			name: "ok",
			req: startAsyncSearchRequest{
				Retention: "60s",
				Query:     testQuery,
				From:      testTimestamp,
				To:        testTimestamp.Add(time.Second),
				WithDocs:  true,
				Size:      100,
				Meta:      meta,
				Histogram: &AsyncSearchRequestHistogram{
					Interval: "1s",
				},
				Aggregations: aggregationTsQueries{
					{
						aggregationQuery: aggregationQuery{
							Field:     "v",
							GroupBy:   "level",
							Func:      afAvg,
							Quantiles: []float64{0.95},
						},
						Interval: "30s",
					},
				},
			},
			want: startAsyncSearchResponse{
				SearchID: testSearchID,
			},
			mockArgs: &mockArgs{
				req: &seqapi.StartAsyncSearchRequest{
					Retention: durationpb.New(60 * time.Second),
					Query:     testQuery,
					From:      timestamppb.New(testTimestamp),
					To:        timestamppb.New(testTimestamp.Add(time.Second)),
					WithDocs:  true,
					Size:      100,
					Hist: &seqapi.StartAsyncSearchRequest_HistQuery{
						Interval: "1s",
					},
					Aggs: []*seqapi.AggregationQuery{
						{
							Field:     "v",
							GroupBy:   "level",
							Func:      seqapi.AggFunc_AGG_FUNC_AVG,
							Quantiles: []float64{0.95},
							Interval:  pointerTo("30s"),
						},
					},
					Meta: meta,
				},
				resp: &seqapi.StartAsyncSearchResponse{
					SearchId: testSearchID,
				},
			},
		},
		{
			name: "err_svc",
			req: startAsyncSearchRequest{
				Retention: "60s",
				Query:     testQuery,
				From:      testTimestamp,
				To:        testTimestamp.Add(time.Second),
				WithDocs:  true,
				Size:      100,
				Meta:      meta,
				Histogram: &AsyncSearchRequestHistogram{
					Interval: "1s",
				},
				Aggregations: aggregationTsQueries{
					{
						aggregationQuery: aggregationQuery{
							Field:     "v",
							GroupBy:   "level",
							Func:      afAvg,
							Quantiles: []float64{0.95},
						},
						Interval: "30s",
					},
				},
			},
			wantErr: true,
			mockArgs: &mockArgs{
				req: &seqapi.StartAsyncSearchRequest{
					Retention: durationpb.New(60 * time.Second),
					Query:     testQuery,
					From:      timestamppb.New(testTimestamp),
					To:        timestamppb.New(testTimestamp.Add(time.Second)),
					WithDocs:  true,
					Size:      100,
					Hist: &seqapi.StartAsyncSearchRequest_HistQuery{
						Interval: "1s",
					},
					Aggs: []*seqapi.AggregationQuery{
						{
							Field:     "v",
							GroupBy:   "level",
							Func:      seqapi.AggFunc_AGG_FUNC_AVG,
							Quantiles: []float64{0.95},
							Interval:  pointerTo("30s"),
						},
					},
					Meta: meta,
				},
				err: errSomethingWrong,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			seqData := test.APITestData{}

			if tt.mockArgs != nil {
				ctrl := gomock.NewController(t)
				svcMock := mock_asyncsearches.NewMockService(ctrl)

				svcMock.EXPECT().
					StartAsyncSearch(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.resp, tt.mockArgs.err).
					Times(1)

				seqData.Mocks.AsyncSearchesSvc = svcMock
			}

			api := setupTestAPI(seqData)

			httputil.DoTestHTTPEx(t, httputil.TestDataHTTPEx[startAsyncSearchRequest, startAsyncSearchResponse]{
				Method:  http.MethodPost,
				Target:  "/seqapi/v1/async_search/start",
				Req:     tt.req,
				Handler: api.serveStartAsyncSearch,
				Want:    tt.want,
				WantErr: tt.wantErr,
			})
		})
	}
}

func TestServeStartAsyncSearch_Disabled(t *testing.T) {
	seqData := test.APITestData{}
	api := setupTestAPI(seqData)

	httputil.DoTestHTTPEx(t, httputil.TestDataHTTPEx[startAsyncSearchRequest, struct{}]{
		Method:  http.MethodPost,
		Target:  "/seqapi/v1/async_search/start",
		Handler: api.serveStartAsyncSearch,
		WantErr: true,
	})
}
