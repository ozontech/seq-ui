package loader

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func writeConfig(t *testing.T, cfg map[string]any) string {
	t.Helper()
	data, err := yaml.Marshal(cfg)
	require.NoError(t, err)

	path := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(path, data, 0o600))
	return path
}

func TestMergeFiles(t *testing.T) {
	tests := []struct {
		name    string
		configs []map[string]any
		want    map[string]any
		wantErr error
	}{
		{
			name: "later_file_overrides_earlier",
			configs: []map[string]any{
				{
					"version": 2,
					"server": map[string]any{
						"http": map[string]any{
							"addr": ":5555",
						},
						"grpc": map[string]any{
							"addr": ":5556",
						},
					},
					"clients": map[string]any{
						"seq_db": map[string]any{
							"test": map[string]any{
								"addrs": []any{
									"seq-db-proxy:9004",
								},
							},
						},
					},
				},
				{
					"version": 2,
					"server": map[string]any{
						"http": map[string]any{
							"addr":         ":6767",
							"read_timeout": "15s",
						},
					},
					"clients": map[string]any{
						"seq_db": map[string]any{
							"test": map[string]any{
								"addrs": []any{
									"seq-db-proxy:9005",
								},
							},
						},
					},
				},
			},
			want: map[string]any{
				"version": 2,
				"server": map[string]any{
					"http": map[string]any{
						"addr":         ":6767",
						"read_timeout": "15s",
					},
					"grpc": map[string]any{
						"addr": ":5556",
					},
				},
				"clients": map[string]any{
					"seq_db": map[string]any{
						"test": map[string]any{
							"addrs": []any{
								"seq-db-proxy:9005",
							},
						},
					},
				},
			},
		},
		{
			name: "different versions",
			configs: []map[string]any{
				{"version": 1},
				{"version": 2},
			},
			wantErr: errors.New("different config versions: 1 and 2"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			paths := make([]string, 0, len(tt.configs))
			for _, cfg := range tt.configs {
				paths = append(paths, writeConfig(t, cfg))
			}

			got, _, err := mergeFiles(paths)
			if tt.wantErr != nil {
				require.Equal(t, tt.wantErr, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
