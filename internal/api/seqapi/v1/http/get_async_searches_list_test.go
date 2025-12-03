package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/api/seqapi/v1/test"
	"github.com/ozontech/seq-ui/internal/app/types"
	mock_seqdb "github.com/ozontech/seq-ui/internal/pkg/client/seqdb/mock"
	mock_repo "github.com/ozontech/seq-ui/internal/pkg/repository/mock"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
)

func TestServeGetAsyncSearchesList(t *testing.T) {
	var (
		mockSearchID1        = "c9a34cf8-4c66-484e-9cc2-42979d848656"
		mockSearchID2        = "9e4c068e-d4f4-4a5d-be27-a6524a70d70d"
		mockUserName1        = "some_user_1"
		mockUserName2        = "some_user_2"
		mockProfileID1 int64 = 1
		mockProfileID2 int64 = 1
		errorMsg             = "some error"

		mockTime = time.Date(2025, 8, 6, 17, 52, 12, 123, time.UTC)
	)

	type mockArgs struct {
		searchIDs []string
		proxyReq  *seqapi.GetAsyncSearchesListRequest
		proxyResp *seqapi.GetAsyncSearchesListResponse
		proxyErr  error

		repoReq  types.GetAsyncSearchesListRequest
		repoResp []types.AsyncSearchInfo
		repoErr  error
	}

	tests := []struct {
		name string

		reqBody      string
		wantRespBody string
		wantStatus   int

		mockArgs *mockArgs
	}{
		{
			name:    "ok_no_filters",
			reqBody: `{}`,
			mockArgs: &mockArgs{
				proxyReq: &seqapi.GetAsyncSearchesListRequest{},
				proxyResp: &seqapi.GetAsyncSearchesListResponse{
					Searches: []*seqapi.GetAsyncSearchesListResponse_ListItem{
						{
							SearchId: mockSearchID1,
							Status:   seqapi.AsyncSearchStatus_ASYNC_SEARCH_STATUS_DONE,
							Request: &seqapi.StartAsyncSearchRequest{
								Retention: durationpb.New(60 * time.Second),
								Query:     "message:error",
								From:      timestamppb.New(mockTime.Add(-15 * time.Minute)),
								To:        timestamppb.New(mockTime),
								WithDocs:  true,
								Size:      100,
							},
							StartedAt: timestamppb.New(mockTime.Add(-30 * time.Second)),
							ExpiresAt: timestamppb.New(mockTime.Add(30 * time.Second)),
							Progress:  1,
							DiskUsage: 512,
							OwnerName: mockUserName1,
							Error:     &errorMsg,
						},
						{
							SearchId: mockSearchID2,
							Status:   seqapi.AsyncSearchStatus_ASYNC_SEARCH_STATUS_CANCELED,
							Request: &seqapi.StartAsyncSearchRequest{
								Retention: durationpb.New(360 * time.Second),
								Query:     "message:error and level:3",
								From:      timestamppb.New(mockTime.Add(-1 * time.Hour)),
								To:        timestamppb.New(mockTime),
								Aggs: []*seqapi.AggregationQuery{
									{
										Field:    "x",
										GroupBy:  "level",
										Func:     seqapi.AggFunc_AGG_FUNC_AVG,
										Interval: pointerTo("30s"),
									},
								},
								Hist: &seqapi.StartAsyncSearchRequest_HistQuery{
									Interval: "1s",
								},
								WithDocs: false,
							},
							StartedAt:  timestamppb.New(mockTime.Add(-60 * time.Second)),
							ExpiresAt:  timestamppb.New(mockTime.Add(300 * time.Second)),
							CanceledAt: timestamppb.New(mockTime),
							Progress:   1,
							DiskUsage:  256,
							OwnerName:  mockUserName2,
						},
					},
				},
				repoReq: types.GetAsyncSearchesListRequest{},
				repoResp: []types.AsyncSearchInfo{
					{
						SearchID:  mockSearchID1,
						OwnerID:   mockProfileID1,
						OwnerName: mockUserName1,
					},
					{
						SearchID:  mockSearchID2,
						OwnerID:   mockProfileID2,
						OwnerName: mockUserName2,
					},
				},
				searchIDs: []string{mockSearchID1, mockSearchID2},
			},
			wantRespBody: `{"searches":[{"search_id":"c9a34cf8-4c66-484e-9cc2-42979d848656","status":"done","request":{"retention":"seconds:60","query":"message:error","from":"2025-08-06T17:37:12.000000123Z","to":"2025-08-06T17:52:12.000000123Z","with_docs":true,"size":100},"started_at":"2025-08-06T17:51:42.000000123Z","expires_at":"2025-08-06T17:52:42.000000123Z","progress":1,"disk_usage":"512","owner_name":"some_user_1","error":"some error"},{"search_id":"9e4c068e-d4f4-4a5d-be27-a6524a70d70d","status":"canceled","request":{"retention":"seconds:360","query":"message:error and level:3","from":"2025-08-06T16:52:12.000000123Z","to":"2025-08-06T17:52:12.000000123Z","aggregations":[{"field":"x","group_by":"level","agg_func":"avg","interval":"30s"}],"histogram":{"interval":"1s"},"with_docs":false,"size":0},"started_at":"2025-08-06T17:51:12.000000123Z","expires_at":"2025-08-06T17:57:12.000000123Z","canceled_at":"2025-08-06T17:52:12.000000123Z","progress":1,"disk_usage":"256","owner_name":"some_user_2"}]}`,
			wantStatus:   http.StatusOK,
		},
		{
			name:    "ok_filters",
			reqBody: `{"limit":10,"offset":20,"status":"done","owner_name":"some_user_1"}`,
			mockArgs: &mockArgs{
				proxyReq: &seqapi.GetAsyncSearchesListRequest{
					Status:    seqapi.AsyncSearchStatus_ASYNC_SEARCH_STATUS_DONE.Enum(),
					OwnerName: &mockUserName1,
					Limit:     10,
					Offset:    20,
				},
				proxyResp: &seqapi.GetAsyncSearchesListResponse{
					Searches: []*seqapi.GetAsyncSearchesListResponse_ListItem{
						{
							SearchId: mockSearchID1,
							Status:   seqapi.AsyncSearchStatus_ASYNC_SEARCH_STATUS_DONE,
							Request: &seqapi.StartAsyncSearchRequest{
								Retention: durationpb.New(60 * time.Second),
								Query:     "message:error",
								From:      timestamppb.New(mockTime.Add(-15 * time.Minute)),
								To:        timestamppb.New(mockTime),
								WithDocs:  true,
								Size:      100,
							},
							StartedAt: timestamppb.New(mockTime.Add(-30 * time.Second)),
							ExpiresAt: timestamppb.New(mockTime.Add(30 * time.Second)),
							Progress:  1,
							DiskUsage: 512,
							OwnerName: mockUserName1,
						},
					},
				},
				repoReq: types.GetAsyncSearchesListRequest{
					Owner: &mockUserName1,
				},
				repoResp: []types.AsyncSearchInfo{
					{
						SearchID:  mockSearchID1,
						OwnerID:   mockProfileID1,
						OwnerName: mockUserName1,
					},
				},
				searchIDs: []string{mockSearchID1},
			},
			wantRespBody: `{"searches":[{"search_id":"c9a34cf8-4c66-484e-9cc2-42979d848656","status":"done","request":{"retention":"seconds:60","query":"message:error","from":"2025-08-06T17:37:12.000000123Z","to":"2025-08-06T17:52:12.000000123Z","with_docs":true,"size":100},"started_at":"2025-08-06T17:51:42.000000123Z","expires_at":"2025-08-06T17:52:42.000000123Z","progress":1,"disk_usage":"512","owner_name":"some_user_1"}]}`,
			wantStatus:   http.StatusOK,
		},
		{
			name:         "err_limit",
			reqBody:      `{"limit":-10,"offset":20}`,
			wantRespBody: `{"message":"invalid request field: 'limit' must be non-negative"}`,
			wantStatus:   http.StatusBadRequest,
		},
		{
			name:         "err_offset",
			reqBody:      `{"limit":10,"offset":-20}`,
			wantRespBody: `{"message":"invalid request field: 'offset' must be non-negative"}`,
			wantStatus:   http.StatusBadRequest,
		},
		{
			name:       "err_invalid_request",
			reqBody:    "invalid-request",
			wantStatus: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			seqData := test.APITestData{}

			if tt.mockArgs != nil {
				ctrl := gomock.NewController(t)

				asyncSearchesRepoMock := mock_repo.NewMockAsyncSearches(ctrl)
				asyncSearchesRepoMock.EXPECT().GetAsyncSearchesList(gomock.Any(), tt.mockArgs.repoReq).
					Return(tt.mockArgs.repoResp, tt.mockArgs.repoErr).Times(1)
				seqData.Mocks.AsyncSearchesRepo = asyncSearchesRepoMock

				seqDbMock := mock_seqdb.NewMockClient(ctrl)
				seqDbMock.EXPECT().GetAsyncSearchesList(gomock.Any(), tt.mockArgs.proxyReq, tt.mockArgs.searchIDs).
					Return(tt.mockArgs.proxyResp, tt.mockArgs.proxyErr).Times(1)
				seqData.Mocks.SeqDB = seqDbMock
			}

			api := initTestAPIWithAsyncSearches(seqData)
			req := httptest.NewRequest(http.MethodPost, "/seqapi/v1/async_search/list", strings.NewReader(tt.reqBody))

			httputil.DoTestHTTP(t, httputil.TestDataHTTP{
				Req:          req,
				Handler:      api.serveGetAsyncSearchesList,
				WantRespBody: tt.wantRespBody,
				WantStatus:   tt.wantStatus,
			})
		})
	}
}

func TestServeGetAsyncSearchesList_Disabled(t *testing.T) {
	seqData := test.APITestData{}
	api := initTestAPI(seqData)
	req := httptest.NewRequest(http.MethodPost, "/seqapi/v1/async_search/list", strings.NewReader("{}"))

	httputil.DoTestHTTP(t, httputil.TestDataHTTP{
		Req:          req,
		Handler:      api.serveGetAsyncSearchesList,
		WantRespBody: `{"message":"async searches disabled"}`,
		WantStatus:   http.StatusBadRequest,
	})
}
