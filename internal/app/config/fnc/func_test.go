// nolint:gofmt
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

			got, gotStatus := tryApply(c.fn)
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

type testIntf interface {
	test()
}

type testImpl struct {
	V string
}

func (i *testImpl) test() {}

func TestApplyFuncs(t *testing.T) {
	t.Parallel()

	type (
		obj struct {
			A int
			B string
			C *obj
		}

		cfg struct {
			NonFuncStr    string
			unsettableStr string

			Str    string
			PtrStr *string

			Strct obj
			Intf  testIntf

			Slc     []string
			SlcPtrs []*string
			SlcObjs []obj

			Arr     [2]string
			ArrPtrs [2]*string
			ArrObjs [1]obj

			M     map[int]string
			MPtrs map[int]*string
			MObjs map[int]obj
		}
	)

	in := cfg{
		NonFuncStr:    "test",
		unsettableStr: "__test__(unset, false)",

		Str:    "__test__(str, false)",
		PtrStr: new(string("__test__(ptrstr, false)")),

		Strct: obj{
			A: 1,
			B: "__test__(strct.b, false)",
			C: &obj{B: "__test__(strct.c.b, false)"},
		},
		Intf: &testImpl{V: "__test__(intf, false)"},

		Slc:     []string{"test", "__test__(slc[1], false)"},
		SlcPtrs: []*string{new(string("__test__(slcptrs[0], false)")), new(string("test"))},
		SlcObjs: []obj{{B: "__test__(slcobjs[0].b, false)"}},

		Arr:     [2]string{"test", "__test__(arr[1], false)"},
		ArrPtrs: [2]*string{new(string("__test__(arrptrs[0], false)")), new(string("test"))},
		ArrObjs: [1]obj{{B: "__test__(arrobjs[0].b, false)"}},

		M: map[int]string{
			1: "__test__(m[1], false)",
			2: "test",
		},
		MPtrs: map[int]*string{
			1: new(string("test")),
			2: new(string("__test__(mptrs[2], false)")),
		},
		MObjs: map[int]obj{
			1: obj{B: "__test__(mobjs[1].b, false)"},
			2: obj{B: "test"},
		},
	}

	want := cfg{
		NonFuncStr:    "test",
		unsettableStr: "__test__(unset, false)",

		Str:    "str",
		PtrStr: new(string("ptrstr")),

		Strct: obj{
			A: 1,
			B: "strct.b",
			C: &obj{B: "strct.c.b"},
		},
		Intf: &testImpl{V: "intf"},

		Slc:     []string{"test", "slc[1]"},
		SlcPtrs: []*string{new(string("slcptrs[0]")), new(string("test"))},
		SlcObjs: []obj{{B: "slcobjs[0].b"}},

		Arr:     [2]string{"test", "arr[1]"},
		ArrPtrs: [2]*string{new(string("arrptrs[0]")), new(string("test"))},
		ArrObjs: [1]obj{{B: "arrobjs[0].b"}},

		M: map[int]string{
			1: "m[1]",
			2: "test",
		},
		MPtrs: map[int]*string{
			1: new(string("test")),
			2: new(string("mptrs[2]")),
		},
		MObjs: map[int]obj{
			1: obj{B: "mobjs[1].b"},
			2: obj{B: "test"},
		},
	}

	ApplyFuncs(&in)
	require.Equal(t, want, in)
}
