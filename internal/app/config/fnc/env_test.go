package fnc

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseEnv(t *testing.T) {
	cases := []struct {
		name string

		args    []string
		wantErr bool
	}{
		{
			name: "ok",
			args: []string{"TEST_ENV"},
		},
		{
			name:    "err_invalid_args_count_1",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "err_invalid_args_count_2",
			args:    []string{"1", "2"},
			wantErr: true,
		},
		{
			name:    "err_empty_arg",
			args:    []string{""},
			wantErr: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			_, gotErr := parseEnv(c.args)
			require.Equal(t, c.wantErr, gotErr != nil)
		})
	}
}

func TestEnvRun(t *testing.T) {
	cases := []struct {
		name string

		args    []string
		envs    [][2]string
		want    string
		wantErr bool
	}{
		{
			name: "ok",
			args: []string{"TEST_ENV"},
			envs: [][2]string{
				{"TEST_ENV", "val"},
			},
			want: "val",
		},
		{
			name:    "err_lookup",
			args:    []string{"TEST_ENV2"},
			wantErr: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			for _, e := range c.envs {
				os.Setenv(e[0], e[1])
			}

			env, err := parseEnv(c.args)
			require.NoError(t, err)

			got, gotErr := env.run()
			require.Equal(t, c.wantErr, gotErr != nil)
			if c.wantErr {
				return
			}

			require.Equal(t, c.want, got)
		})
	}
}
