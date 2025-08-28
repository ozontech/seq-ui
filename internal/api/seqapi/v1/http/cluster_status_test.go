package http

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/api/seqapi/v1/test"
	mock_seqdb "github.com/ozontech/seq-ui/internal/pkg/client/seqdb/mock"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestStatus(t *testing.T) {
	type mockArgs struct {
		resp *seqapi.StatusResponse
		err  error
	}

	type testCase struct {
		name string

		wantRespBody string
		wantStatus   int

		mockArgs mockArgs
	}

	someMoment := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []testCase{
		{
			name: "ok",
			mockArgs: mockArgs{
				resp: &seqapi.StatusResponse{
					NumberOfStores:    1,
					OldestStorageTime: timestamppb.New(someMoment),
					Stores: []*seqapi.StoreStatus{
						{
							Host:   "host-0",
							Values: &seqapi.StoreStatusValues{OldestTime: timestamppb.New(someMoment)},
						},
					},
				},
			},
			wantRespBody: `{"oldest_storage_time":"2020-01-01T00:00:00Z","number_of_stores":1,"stores":[{"host":"host-0","values":{"oldest_time":"2020-01-01T00:00:00Z"}}]}`,
			wantStatus:   http.StatusOK,
		},
		{
			name: "err_client",
			mockArgs: mockArgs{
				err: errors.New("client error"),
			},
			wantStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			seqData := test.APITestData{}

			seqDbMock := mock_seqdb.NewMockClient(ctrl)
			seqDbMock.EXPECT().Status(gomock.Any(), gomock.Any()).
				Return(tt.mockArgs.resp, tt.mockArgs.err).Times(1)
			seqData.Mocks.SeqDB = seqDbMock

			api := initTestAPI(seqData)
			req := httptest.NewRequest(http.MethodGet, "/seqapi/v1/status", http.NoBody)

			httputil.DoTestHTTP(t, httputil.TestDataHTTP{
				Req:          req,
				Handler:      api.serveStatus,
				WantRespBody: tt.wantRespBody,
				WantStatus:   tt.wantStatus,
			})
		})
	}
}
