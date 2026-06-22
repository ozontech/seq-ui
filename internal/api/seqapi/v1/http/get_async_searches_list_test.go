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

func TestServeGetAsyncSearchesList(t *testing.T) {
	statusDone := asyncSearchStatus("done")

	type mockArgs struct {
		req  *seqapi.GetAsyncSearchesListRequest
		resp *seqapi.GetAsyncSearchesListResponse
		err  error
	}

	tests := []struct {
		name string

		req     getAsyncSearchesListRequest
		want    getAsyncSearchesListResponse
		wantErr bool

		mockArgs *mockArgs
	}{
		{
			name: "ok_no_filters",
			req:  getAsyncSearchesListRequest{},
			want: getAsyncSearchesListResponseFromProto(&seqapi.GetAsyncSearchesListResponse{
				Searches: []*seqapi.GetAsyncSearchesListResponse_ListItem{
					{
						SearchId: mockSearchID,
						Status:   seqapi.AsyncSearchStatus_ASYNC_SEARCH_STATUS_DONE,
						Request: &seqapi.StartAsyncSearchRequest{
							Retention: durationpb.New(60 * time.Second),
							Query:     "message:error",
							From:      timestamppb.New(someMoment.Add(-15 * time.Minute)),
							To:        timestamppb.New(someMoment),
							WithDocs:  true,
							Size:      100,
						},
						StartedAt: timestamppb.New(someMoment.Add(-30 * time.Second)),
						ExpiresAt: timestamppb.New(someMoment.Add(30 * time.Second)),
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
							From:      timestamppb.New(someMoment.Add(-1 * time.Hour)),
							To:        timestamppb.New(someMoment),
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
						StartedAt:  timestamppb.New(someMoment.Add(-60 * time.Second)),
						ExpiresAt:  timestamppb.New(someMoment.Add(300 * time.Second)),
						CanceledAt: timestamppb.New(someMoment),
						Progress:   1,
						DiskUsage:  256,
						OwnerName:  mockUserName2,
					},
				},
				Error: &seqapi.Error{
					Code: seqapi.ErrorCode_ERROR_CODE_NO,
				},
			}),
			mockArgs: &mockArgs{
				req: &seqapi.GetAsyncSearchesListRequest{},
				resp: &seqapi.GetAsyncSearchesListResponse{
					Searches: []*seqapi.GetAsyncSearchesListResponse_ListItem{
						{
							SearchId: mockSearchID,
							Status:   seqapi.AsyncSearchStatus_ASYNC_SEARCH_STATUS_DONE,
							Request: &seqapi.StartAsyncSearchRequest{
								Retention: durationpb.New(60 * time.Second),
								Query:     "message:error",
								From:      timestamppb.New(someMoment.Add(-15 * time.Minute)),
								To:        timestamppb.New(someMoment),
								WithDocs:  true,
								Size:      100,
							},
							StartedAt: timestamppb.New(someMoment.Add(-30 * time.Second)),
							ExpiresAt: timestamppb.New(someMoment.Add(30 * time.Second)),
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
								From:      timestamppb.New(someMoment.Add(-1 * time.Hour)),
								To:        timestamppb.New(someMoment),
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
							StartedAt:  timestamppb.New(someMoment.Add(-60 * time.Second)),
							ExpiresAt:  timestamppb.New(someMoment.Add(300 * time.Second)),
							CanceledAt: timestamppb.New(someMoment),
							Progress:   1,
							DiskUsage:  256,
							OwnerName:  mockUserName2,
						},
					},
					Error: &seqapi.Error{
						Code: seqapi.ErrorCode_ERROR_CODE_NO,
					},
				},
			},
		},
		{
			name: "ok_filters",
			req: getAsyncSearchesListRequest{
				Status: &statusDone,
				Limit:  10,
				Offset: 20,
				Owner:  &mockUserName1,
			},
			want: getAsyncSearchesListResponseFromProto(&seqapi.GetAsyncSearchesListResponse{
				Searches: []*seqapi.GetAsyncSearchesListResponse_ListItem{
					{
						SearchId: mockSearchID,
						Status:   seqapi.AsyncSearchStatus_ASYNC_SEARCH_STATUS_DONE,
						Request: &seqapi.StartAsyncSearchRequest{
							Retention: durationpb.New(60 * time.Second),
							Query:     "message:error",
							From:      timestamppb.New(someMoment.Add(-15 * time.Minute)),
							To:        timestamppb.New(someMoment),
							WithDocs:  true,
							Size:      100,
						},
						StartedAt: timestamppb.New(someMoment.Add(-30 * time.Second)),
						ExpiresAt: timestamppb.New(someMoment.Add(30 * time.Second)),
						Progress:  1,
						DiskUsage: 512,
						OwnerName: mockUserName1,
					},
				},
				Error: &seqapi.Error{
					Code: seqapi.ErrorCode_ERROR_CODE_NO,
				},
			}),
			mockArgs: &mockArgs{
				req: &seqapi.GetAsyncSearchesListRequest{
					Status:    seqapi.AsyncSearchStatus_ASYNC_SEARCH_STATUS_DONE.Enum(),
					OwnerName: &mockUserName1,
					Limit:     10,
					Offset:    20,
				},
				resp: &seqapi.GetAsyncSearchesListResponse{
					Searches: []*seqapi.GetAsyncSearchesListResponse_ListItem{
						{
							SearchId: mockSearchID,
							Status:   seqapi.AsyncSearchStatus_ASYNC_SEARCH_STATUS_DONE,
							Request: &seqapi.StartAsyncSearchRequest{
								Retention: durationpb.New(60 * time.Second),
								Query:     "message:error",
								From:      timestamppb.New(someMoment.Add(-15 * time.Minute)),
								To:        timestamppb.New(someMoment),
								WithDocs:  true,
								Size:      100,
							},
							StartedAt: timestamppb.New(someMoment.Add(-30 * time.Second)),
							ExpiresAt: timestamppb.New(someMoment.Add(30 * time.Second)),
							Progress:  1,
							DiskUsage: 512,
							OwnerName: mockUserName1,
						},
					},
					Error: &seqapi.Error{
						Code: seqapi.ErrorCode_ERROR_CODE_NO,
					},
				},
			},
		},
		{
			name: "partial_response",
			req:  getAsyncSearchesListRequest{},
			want: getAsyncSearchesListResponseFromProto(&seqapi.GetAsyncSearchesListResponse{
				Searches: []*seqapi.GetAsyncSearchesListResponse_ListItem{
					{
						SearchId: mockSearchID,
						Status:   seqapi.AsyncSearchStatus_ASYNC_SEARCH_STATUS_DONE,
						Request: &seqapi.StartAsyncSearchRequest{
							Retention: durationpb.New(60 * time.Second),
							Query:     "message:error",
							From:      timestamppb.New(someMoment.Add(-15 * time.Minute)),
							To:        timestamppb.New(someMoment),
							WithDocs:  true,
							Size:      100,
						},
						StartedAt: timestamppb.New(someMoment.Add(-30 * time.Second)),
						ExpiresAt: timestamppb.New(someMoment.Add(30 * time.Second)),
						Progress:  1,
						DiskUsage: 512,
						OwnerName: mockUserName1,
					},
				},
				Error: &seqapi.Error{
					Code:    seqapi.ErrorCode_ERROR_CODE_PARTIAL_RESPONSE,
					Message: "partial response",
				},
			}),
			mockArgs: &mockArgs{
				req: &seqapi.GetAsyncSearchesListRequest{},
				resp: &seqapi.GetAsyncSearchesListResponse{
					Searches: []*seqapi.GetAsyncSearchesListResponse_ListItem{
						{
							SearchId: mockSearchID,
							Status:   seqapi.AsyncSearchStatus_ASYNC_SEARCH_STATUS_DONE,
							Request: &seqapi.StartAsyncSearchRequest{
								Retention: durationpb.New(60 * time.Second),
								Query:     "message:error",
								From:      timestamppb.New(someMoment.Add(-15 * time.Minute)),
								To:        timestamppb.New(someMoment),
								WithDocs:  true,
								Size:      100,
							},
							StartedAt: timestamppb.New(someMoment.Add(-30 * time.Second)),
							ExpiresAt: timestamppb.New(someMoment.Add(30 * time.Second)),
							Progress:  1,
							DiskUsage: 512,
							OwnerName: mockUserName1,
						},
					},
					Error: &seqapi.Error{
						Code:    seqapi.ErrorCode_ERROR_CODE_PARTIAL_RESPONSE,
						Message: "partial response",
					},
				},
			},
		},
		{
			name: "err_limit",
			req: getAsyncSearchesListRequest{
				Limit:  -10,
				Offset: 20,
			},
			wantErr: true,
		},
		{
			name: "err_offset",
			req: getAsyncSearchesListRequest{
				Limit:  10,
				Offset: -20,
			},
			wantErr: true,
		},
		{
			name: "query_too_long",
			req:  getAsyncSearchesListRequest{},
			want: getAsyncSearchesListResponseFromProto(&seqapi.GetAsyncSearchesListResponse{
				Searches: []*seqapi.GetAsyncSearchesListResponse_ListItem{
					{
						SearchId: mockSearchID,
						Status:   seqapi.AsyncSearchStatus_ASYNC_SEARCH_STATUS_DONE,
						Request: &seqapi.StartAsyncSearchRequest{
							Retention: durationpb.New(60 * time.Second),
							Query:     tooLongQuery,
							From:      timestamppb.New(someMoment.Add(-15 * time.Minute)),
							To:        timestamppb.New(someMoment),
							WithDocs:  true,
							Size:      100,
						},
						StartedAt: timestamppb.New(someMoment.Add(-30 * time.Second)),
						ExpiresAt: timestamppb.New(someMoment.Add(30 * time.Second)),
						Progress:  1,
						DiskUsage: 512,
						OwnerName: mockUserName1,
					},
				},
				Error: &seqapi.Error{
					Code:    seqapi.ErrorCode_ERROR_CODE_PARTIAL_RESPONSE,
					Message: "partial response",
				},
			}),
			mockArgs: &mockArgs{
				req: &seqapi.GetAsyncSearchesListRequest{},
				resp: &seqapi.GetAsyncSearchesListResponse{
					Searches: []*seqapi.GetAsyncSearchesListResponse_ListItem{
						{
							SearchId: mockSearchID,
							Status:   seqapi.AsyncSearchStatus_ASYNC_SEARCH_STATUS_DONE,
							Request: &seqapi.StartAsyncSearchRequest{
								Retention: durationpb.New(60 * time.Second),
								Query:     tooLongQuery,
								From:      timestamppb.New(someMoment.Add(-15 * time.Minute)),
								To:        timestamppb.New(someMoment),
								WithDocs:  true,
								Size:      100,
							},
							StartedAt: timestamppb.New(someMoment.Add(-30 * time.Second)),
							ExpiresAt: timestamppb.New(someMoment.Add(30 * time.Second)),
							Progress:  1,
							DiskUsage: 512,
							OwnerName: mockUserName1,
						},
					},
					Error: &seqapi.Error{
						Code:    seqapi.ErrorCode_ERROR_CODE_PARTIAL_RESPONSE,
						Message: "partial response",
					},
				},
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
					GetAsyncSearchesList(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.resp, tt.mockArgs.err).
					Times(1)
				seqData.Mocks.AsyncSearchesSvc = svcMock
			}

			api := setupAPIWithAsyncSearches(seqData)

			httputil.DoTestHTTPEx(t, httputil.TestDataHTTPEx[getAsyncSearchesListRequest, getAsyncSearchesListResponse]{
				Method:  http.MethodPost,
				Target:  "/seqapi/v1/async_search/list",
				Req:     tt.req,
				Handler: api.serveGetAsyncSearchesList,
				Want:    tt.want,
				WantErr: tt.wantErr,
			})
		})
	}
}

func TestServeGetAsyncSearchesList_Disabled(t *testing.T) {
	seqData := test.APITestData{}
	api := setupAPI(seqData)

	httputil.DoTestHTTPEx(t, httputil.TestDataHTTPEx[getAsyncSearchesListRequest, struct{}]{
		Method:  http.MethodPost,
		Target:  "/seqapi/v1/async_search/list",
		Handler: api.serveGetAsyncSearchesList,
		WantErr: true,
	})
}
