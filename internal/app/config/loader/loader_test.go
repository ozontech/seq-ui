package loader

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFromFiles(t *testing.T) {
	base := map[string]any{
		"version": 2,
		"server": map[string]any{
			"http": map[string]any{
				"addr": ":5555",
			},
		},
		"handlers": map[string]any{
			"seq_api": map[string]any{
				"options": map[string]any{
					"seq_db_id": "test",
					"limits": map[string]any{
						"max_search_limit": 100,
						"max_export_limit": 100,
					},
				},
			},
		},
	}
	override := map[string]any{
		"version": 2,
		"server": map[string]any{
			"http": map[string]any{
				"addr": ":6767",
			},
		},
		"handlers": map[string]any{
			"seq_api": map[string]any{
				"options": map[string]any{
					"limits": map[string]any{
						"max_search_limit": 10,
					},
				},
			},
		},
		"clients": map[string]any{
			"seq_db": map[string]any{
				"test": map[string]any{
					"addrs": []any{
						"seq-db-proxy:9005",
					},
					"timeout": "15s",
				},
			},
		},
	}
	expected := map[string]any{
		"version": 2,
		"server": map[string]any{
			"http": map[string]any{
				"addr": ":6767",
			},
		},
		"handlers": map[string]any{
			"seq_api": map[string]any{
				"options": map[string]any{
					"seq_db_id": "test",
					"limits": map[string]any{
						"max_search_limit": 10,
						"max_export_limit": 100,
					},
				},
			},
		},
		"clients": map[string]any{
			"seq_db": map[string]any{
				"test": map[string]any{
					"addrs": []any{
						"seq-db-proxy:9005",
					},
					"timeout": "15s",
				},
			},
		},
	}

	got, err := FromFiles([]string{writeConfig(t, base), writeConfig(t, override)})
	require.NoError(t, err)

	want, err := FromFiles([]string{writeConfig(t, expected)})
	require.NoError(t, err)

	require.Equal(t, want, got)
}
