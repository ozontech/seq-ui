package http

import (
	"fmt"
	"net/http"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/api/seqapi/v1/test"
	mock_asyncsearches "github.com/ozontech/seq-ui/internal/pkg/service/async_searches/mock"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
)

func TestServeCancelAsyncSearch(t *testing.T) {
	type mockArgs struct {
		req  *seqapi.CancelAsyncSearchRequest
		resp *seqapi.CancelAsyncSearchResponse
		err  error
	}

	tests := []struct {
		name string

		searchID string
		wantErr  bool
		noResp   bool

		mockArgs *mockArgs
	}{
		{
			name:     "ok",
			searchID: mockSearchID,
			noResp:   true,
			mockArgs: &mockArgs{
				req: &seqapi.CancelAsyncSearchRequest{
					SearchId: mockSearchID,
				},
				resp: &seqapi.CancelAsyncSearchResponse{},
			},
		},
		{
			name:     "invalid_id",
			searchID: "some invalid id",
			wantErr:  true,
		},
		{
			name:     "err_svc",
			searchID: mockSearchID,
			wantErr:  true,
			mockArgs: &mockArgs{
				req: &seqapi.CancelAsyncSearchRequest{
					SearchId: mockSearchID,
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

				if tt.mockArgs.req != nil {
					svcMock.EXPECT().
						CancelAsyncSearch(gomock.Any(), tt.mockArgs.req).
						Return(tt.mockArgs.resp, tt.mockArgs.err).
						Times(1)
				}

				seqData.Mocks.AsyncSearchesSvc = svcMock
			}

			api := setupAPIWithAsyncSearches(seqData)

			httputil.DoTestHTTPEx(t, httputil.TestDataHTTPEx[struct{}, struct{}]{
				Method:  http.MethodPost,
				Target:  fmt.Sprintf("/seqapi/v1/async_search/%s/cancel", mockSearchID),
				Handler: withAsyncSearchID(api.serveCancelAsyncSearch, tt.searchID),
				WantErr: tt.wantErr,
				NoResp:  tt.noResp,
			})
		})
	}
}

func TestServeCancelAsyncSearch_Disabled(t *testing.T) {
	seqData := test.APITestData{}
	api := setupAPI(seqData)

	httputil.DoTestHTTPEx(t, httputil.TestDataHTTPEx[struct{}, struct{}]{
		Method:  http.MethodPost,
		Target:  "/seqapi/v1/async_search/c9a34cf8-4c66-484e-9cc2-42979d848656/cancel",
		Handler: api.serveCancelAsyncSearch,
		WantErr: true,
	})
}
