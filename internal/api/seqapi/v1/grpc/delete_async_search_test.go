package grpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ozontech/seq-ui/internal/api/seqapi/v1/test"
	"github.com/ozontech/seq-ui/internal/app/types"
	mock_asyncsearches "github.com/ozontech/seq-ui/internal/pkg/service/async_searches/mock"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
)

func TestDeleteAsyncSearch(t *testing.T) {
	type mockArgs struct {
		req  *seqapi.DeleteAsyncSearchRequest
		resp *seqapi.DeleteAsyncSearchResponse
		err  error
	}

	tests := []struct {
		name string

		req      *seqapi.DeleteAsyncSearchRequest
		want     *seqapi.DeleteAsyncSearchResponse
		wantCode codes.Code

		mockArgs *mockArgs
	}{
		{
			name: "ok",
			req: &seqapi.DeleteAsyncSearchRequest{
				SearchId: testSearchID,
			},
			want: &seqapi.DeleteAsyncSearchResponse{},
			mockArgs: &mockArgs{
				req: &seqapi.DeleteAsyncSearchRequest{
					SearchId: testSearchID,
				},
				resp: &seqapi.DeleteAsyncSearchResponse{},
			},
		},
		{
			name: "invalid_id",
			req: &seqapi.DeleteAsyncSearchRequest{
				SearchId: "some_invalid_id",
			},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "err_svc",
			req: &seqapi.DeleteAsyncSearchRequest{
				SearchId: testSearchID,
			},
			wantCode: codes.Internal,
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

			ctrl := gomock.NewController(t)
			svcMock := mock_asyncsearches.NewMockService(ctrl)

			seqData := test.APITestData{}
			seqData.Mocks.AsyncSearchesSvc = svcMock

			if tt.mockArgs != nil {
				svcMock.EXPECT().
					DeleteAsyncSearch(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.resp, tt.mockArgs.err).
					Times(1)
			}

			api := setupTestAPI(seqData)
			got, err := api.DeleteAsyncSearch(context.Background(), tt.req)

			require.Equal(t, tt.wantCode, status.Code(err))
			if tt.wantCode != codes.OK {
				return
			}
			require.Equal(t, tt.want, got)
		})
	}
}

func TestServeDeleteAsyncSearch_Disabled(t *testing.T) {
	seqData := test.APITestData{}
	api := setupTestAPI(seqData)

	_, err := api.DeleteAsyncSearch(context.Background(), &seqapi.DeleteAsyncSearchRequest{SearchId: testSearchID})
	require.Error(t, err)
	require.Equal(t, status.Error(codes.Unimplemented, types.ErrAsyncSearchesDisabled.Error()), err)
}
