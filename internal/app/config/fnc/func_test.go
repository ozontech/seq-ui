package fnc

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTryApply(t *testing.T) {
	cases := []struct {
		name string

		fn         string
		want       string
		wantStatus bool
	}{
		{
			name:       "ok",
			fn:         "__test__(tst, false)",
			want:       "tst",
			wantStatus: true,
		},
		{
			name:       "ok_escaped",
			fn:         `\__test__()`,
			want:       "__test__()",
			wantStatus: true,
		},
		{
			name:       "ok_unknown",
			fn:         `unknownfunc()`,
			wantStatus: false,
		},
		{
			name:       "err_parse",
			fn:         "__test__(",
			wantStatus: false,
		},
		{
			name:       "err_run",
			fn:         "__test__(tst, true)",
			wantStatus: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			got, gotStatus := TryApply(c.fn)
			require.Equal(t, c.want, got)
			require.Equal(t, c.wantStatus, gotStatus)
		})
	}
}

func TestParse(t *testing.T) {
	cases := []struct {
		name string

		fn      string
		want    fnc
		wantErr bool
	}{
		{
			name: "ok_not_fn",
			fn:   "test",
			want: nil,
		},
		{
			name:    "ok_escaped",
			fn:      `\__test__()`,
			wantErr: true, // custom errEscaped, but valid case
		},
		{
			name: "ok_valid_fn",
			fn:   "__test__(1,true)",
			want: &test{want: "1", wantErr: true},
		},
		{
			name: "ok_unknown_fn",
			fn:   `unknownfn()`,
			want: nil,
		},
		{
			name:    "err_missing_parenthesis",
			fn:      `__test__(`,
			wantErr: true,
		},
		{
			name:    "err_new_fn",
			fn:      `__test__()`,
			wantErr: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			got, gotErr := parse(c.fn)
			require.Equal(t, c.wantErr, gotErr != nil)
			if c.wantErr {
				return
			}

			require.Equal(t, c.want, got)
		})
	}
}
