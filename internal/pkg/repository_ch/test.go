package repositorych

import (
	"testing"
	"time"

	mock "github.com/ozontech/seq-ui/internal/pkg/repository_ch/mock"
	"go.uber.org/mock/gomock"
)

func fakeNow(now time.Time) func() time.Time {
	return func() time.Time {
		return now
	}
}

type mockConnRow struct {
	query string
	args  []any

	scanErr error
}

type mockRowsCount struct {
	count   int
	scanErr error
}

type mockRowsScan struct {
	scanFns []func(...any) error
	scanErr bool
}

type mockRowsScanStruct struct {
	scanStructFns []func(any) error
	scanErr       bool
}

type mockConnRows struct {
	query string
	args  []any

	rows any
	err  error
}

func initMockConnRow(t *testing.T, params ...*mockConnRow) *mock.MockConn {
	ctrl := gomock.NewController(t)
	mc := mock.NewMockConn(ctrl)

	for _, p := range params {
		if p == nil {
			continue
		}

		mockedRow := mock.NewMockRow(ctrl)
		mockedRow.EXPECT().Scan(gomock.Any()).Return(p.scanErr)

		if p.query == "" {
			mc.EXPECT().
				QueryRow(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(mockedRow).
				Times(1)
		} else {
			mc.EXPECT().
				QueryRow(gomock.Any(), p.query, p.args...).
				Return(mockedRow).
				Times(1)
		}
	}

	return mc
}

func initMockConnRows(t *testing.T, params ...*mockConnRows) *mock.MockConn {
	ctrl := gomock.NewController(t)
	mc := mock.NewMockConn(ctrl)

	for _, p := range params {
		if p == nil {
			continue
		}

		mockedRows := mock.NewMockRows(ctrl)

		if p.rows != nil {
			switch rows := (p.rows).(type) {
			case *mockRowsCount:
				times := rows.count
				if rows.scanErr != nil {
					times = 1
				}

				mockedRows.EXPECT().Next().Return(true).Times(times)
				mockedRows.EXPECT().Scan(gomock.Any()).Return(rows.scanErr).Times(times)

				if rows.scanErr == nil {
					mockedRows.EXPECT().Next().Return(false).Times(1)
				}
			case *mockRowsScan:
				for _, fn := range rows.scanFns {
					mockedRows.EXPECT().Next().Return(true)
					mockedRows.EXPECT().Scan(gomock.Any()).DoAndReturn(fn)
				}

				if !rows.scanErr {
					mockedRows.EXPECT().Next().Return(false)
				}

			case *mockRowsScanStruct:
				for _, fn := range rows.scanStructFns {
					mockedRows.EXPECT().Next().Return(true)
					mockedRows.EXPECT().ScanStruct(gomock.Any()).DoAndReturn(fn)
				}

				if !rows.scanErr {
					mockedRows.EXPECT().Next().Return(false)
				}
			}
		}

		if p.query == "" {
			mc.EXPECT().
				Query(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(mockedRows, p.err).
				Times(1)
		} else {
			mc.EXPECT().
				Query(gomock.Any(), p.query, p.args...).
				Return(mockedRows, p.err).
				Times(1)
		}
	}

	return mc
}
