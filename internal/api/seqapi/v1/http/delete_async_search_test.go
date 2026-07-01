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

func TestServeDeleteAsyncSearch(t *testing.T) {
	type mockArgs struct {
		req  *seqapi.DeleteAsyncSearchRequest
		resp *seqapi.DeleteAsyncSearchResponse
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
			searchID: testSearchID,
			noResp:   true,
			mockArgs: &mockArgs{
				req: &seqapi.DeleteAsyncSearchRequest{
					SearchId: testSearchID,
				},
				resp: &seqapi.DeleteAsyncSearchResponse{},
			},
		},
		{
			name:     "invalid_id",
			searchID: "some invalid id",
			wantErr:  true,
		},
		{
			name:     "err_svc",
			searchID: testSearchID,
			wantErr:  true,
			mockArgs: &mockArgs{
				req: &seqapi.DeleteAsyncSearchRequest{
					SearchId: testSearchID,
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
					DeleteAsyncSearch(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.resp, tt.mockArgs.err).
					Times(1)

				seqData.Mocks.AsyncSearchesSvc = svcMock
			}

			api := setupTestAPI(seqData)

			httputil.DoTestHTTPEx(t, httputil.TestDataHTTPEx[struct{}, struct{}]{
				Method:  http.MethodDelete,
				Target:  fmt.Sprintf("/seqapi/v1/async_search/%s", testSearchID),
				Handler: withQueryParamID(api.serveDeleteAsyncSearch, tt.searchID),
				WantErr: tt.wantErr,
				NoResp:  tt.noResp,
			})
		})
	}
}

func TestServeDeleteAsyncSearch_Disabled(t *testing.T) {
	seqData := test.APITestData{}
	api := setupTestAPI(seqData)

	httputil.DoTestHTTPEx(t, httputil.TestDataHTTPEx[struct{}, struct{}]{
		Method:  http.MethodDelete,
		Target:  fmt.Sprintf("/seqapi/v1/async_search/%s", testSearchID),
		Handler: api.serveDeleteAsyncSearch,
		WantErr: true,
	})
}
