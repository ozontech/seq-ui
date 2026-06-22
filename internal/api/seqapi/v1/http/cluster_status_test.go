package http

import (
	"net/http"
	"testing"

	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/api/seqapi/v1/test"
	mock_seqdb "github.com/ozontech/seq-ui/internal/pkg/client/seqdb/mock"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
)

func TestStatus(t *testing.T) {
	type mockArgs struct {
		resp *seqapi.StatusResponse
		err  error
	}

	tests := []struct {
		name string

		want    statusResponse
		wantErr bool

		mockArgs *mockArgs
	}{
		{
			name: "ok",
			want: statusResponse{
				OldestStorageTime: &testSomeMoment,
				NumberOfStores:    1,
				Stores: []storeStatus{
					{
						Host:   "host-0",
						Values: &storeStatusValues{OldestTime: &testSomeMoment},
					},
				},
			},
			mockArgs: &mockArgs{
				resp: &seqapi.StatusResponse{
					NumberOfStores:    1,
					OldestStorageTime: timestamppb.New(testSomeMoment),
					Stores: []*seqapi.StoreStatus{
						{
							Host:   "host-0",
							Values: &seqapi.StoreStatusValues{OldestTime: timestamppb.New(testSomeMoment)},
						},
					},
				},
			},
		},
		{
			name:    "err_client",
			wantErr: true,
			mockArgs: &mockArgs{
				err: errSomethingWrong,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			seqData := test.APITestData{}
			ctrl := gomock.NewController(t)
			seqDbMock := mock_seqdb.NewMockClient(ctrl)

			seqDbMock.EXPECT().
				Status(gomock.Any(), gomock.Any()).
				Return(tt.mockArgs.resp, tt.mockArgs.err).
				Times(1)

			seqData.Mocks.SeqDB = seqDbMock
			api := setupTestAPI(seqData)

			httputil.DoTestHTTPEx(t, httputil.TestDataHTTPEx[struct{}, statusResponse]{
				Method:  http.MethodGet,
				Target:  "/seqapi/v1/status",
				Handler: api.serveStatus,
				Want:    tt.want,
				WantErr: tt.wantErr,
			})
		})
	}
}
