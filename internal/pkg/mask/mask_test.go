package mask

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ozontech/seq-ui/internal/app/config"
)

func TestMaskApply(t *testing.T) {
	tests := []struct {
		name string

		cfg                 config.Mask
		globalProcessFields []string
		globalIgnoreFields  []string

		input map[string]string
		want  map[string]string
	}{
		{
			name: "full_mask",
			cfg: config.Mask{
				Re:     `(\d{3})-(\d{3})-(\d{4})`,
				Groups: []int{0},
				Mode:   config.MaskModeMask,
			},
			input: map[string]string{
				"f1": "my number: 123_456_7890;",
				"f2": "my number: 123-456-7890;",
			},
			want: map[string]string{
				"f1": "my number: 123_456_7890;",
				"f2": "my number: ************;",
			},
		},
		{
			name: "full_replace",
			cfg: config.Mask{
				Re:          `(\d{3})-(\d{3})-(\d{4})`,
				Groups:      []int{0},
				Mode:        config.MaskModeReplace,
				ReplaceWord: "<number>",
			},
			input: map[string]string{
				"f1": "my number: 123_456_7890;",
				"f2": "my number: 123-456-7890;",
			},
			want: map[string]string{
				"f1": "my number: 123_456_7890;",
				"f2": "my number: <number>;",
			},
		},
		{
			name: "full_cut",
			cfg: config.Mask{
				Re:     `(\d{3})-(\d{3})-(\d{4})`,
				Groups: []int{0},
				Mode:   config.MaskModeCut,
			},
			input: map[string]string{
				"f1": "my number: 123_456_7890;",
				"f2": "my number: 123-456-7890;",
			},
			want: map[string]string{
				"f1": "my number: 123_456_7890;",
				"f2": "my number: ;",
			},
		},
		{
			name: "groups",
			cfg: config.Mask{
				Re:     `(\d{3})-(\d{3})-(\d{4})`,
				Groups: []int{1, 3},
				Mode:   config.MaskModeMask,
			},
			input: map[string]string{
				"f1": "my number: 123_456_7890;",
				"f2": "my number: 123-456-7890;",
			},
			want: map[string]string{
				"f1": "my number: 123_456_7890;",
				"f2": "my number: ***-456-****;",
			},
		},
		{
			name: "process_local",
			cfg: config.Mask{
				Re:            `(\d{3})-(\d{3})-(\d{4})`,
				Groups:        []int{0},
				Mode:          config.MaskModeMask,
				ProcessFields: []string{"f1"},
			},
			globalProcessFields: []string{"f2"},
			input: map[string]string{
				"f1": "my number: 098-765-4321;",
				"f2": "my number: 123-456-7890;",
			},
			want: map[string]string{
				"f1": "my number: ************;",
				"f2": "my number: 123-456-7890;",
			},
		},
		{
			name: "process_global",
			cfg: config.Mask{
				Re:     `(\d{3})-(\d{3})-(\d{4})`,
				Groups: []int{0},
				Mode:   config.MaskModeMask,
			},
			globalProcessFields: []string{"f2"},
			input: map[string]string{
				"f1": "my number: 098-765-4321;",
				"f2": "my number: 123-456-7890;",
			},
			want: map[string]string{
				"f1": "my number: 098-765-4321;",
				"f2": "my number: ************;",
			},
		},
		{
			name: "ignore_local",
			cfg: config.Mask{
				Re:           `(\d{3})-(\d{3})-(\d{4})`,
				Groups:       []int{0},
				Mode:         config.MaskModeMask,
				IgnoreFields: []string{"f1"},
			},
			globalIgnoreFields: []string{"f2"},
			input: map[string]string{
				"f1": "my number: 098-765-4321;",
				"f2": "my number: 123-456-7890;",
			},
			want: map[string]string{
				"f1": "my number: 098-765-4321;",
				"f2": "my number: ************;",
			},
		},
		{
			name: "ignore_global",
			cfg: config.Mask{
				Re:     `(\d{3})-(\d{3})-(\d{4})`,
				Groups: []int{0},
				Mode:   config.MaskModeMask,
			},
			globalIgnoreFields: []string{"f2"},
			input: map[string]string{
				"f1": "my number: 098-765-4321;",
				"f2": "my number: 123-456-7890;",
			},
			want: map[string]string{
				"f1": "my number: ************;",
				"f2": "my number: 123-456-7890;",
			},
		},
		{
			name: "no_groups",
			cfg: config.Mask{
				Re:   `\d{3}-\d{3}-\d{4}`,
				Mode: config.MaskModeMask,
			},
			input: map[string]string{
				"f1": "my number: 123_456_7890;",
				"f2": "my number: 123-456-7890;",
			},
			want: map[string]string{
				"f1": "my number: 123_456_7890;",
				"f2": "my number: ************;",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var (
				globalFields *maskFields
				err          error
			)
			if len(tt.globalProcessFields) > 0 || len(tt.globalIgnoreFields) > 0 {
				globalFields, err = parseFields(tt.globalProcessFields, tt.globalIgnoreFields)
				require.NoError(t, err)
			}

			mask, err := compileMask(tt.cfg, globalFields)
			require.NoError(t, err)

			in := tt.input
			mask.apply(in)

			for k, v := range tt.want {
				gotV, ok := in[k]
				require.True(t, ok)
				assert.Equal(t, v, gotV, "wrong value with key %q", k)
			}
		})
	}
}
