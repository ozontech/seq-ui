package mask

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ozontech/seq-ui/internal/app/config"
)

func TestMaskerMask(t *testing.T) {
	tests := []struct {
		name string

		cfg *config.Masking

		input map[string]string
		want  map[string]string
	}{
		{
			name: "multiple_full_mask",
			cfg: &config.Masking{
				Masks: []config.Mask{
					{
						Re:   `(\d{3})-(\d{3})-(\d{4})`,
						Mode: config.MaskModeMask,
					},
					{
						Re:   `(test)`,
						Mode: config.MaskModeMask,
					},
				},
			},
			input: map[string]string{
				"f1": "my number: 123_456_7890 test;",
				"f2": "my number: 123-456-7890;",
				"f3": "my number: 123-456-7890 test;",
			},
			want: map[string]string{
				"f1": "my number: 123_456_7890 ****;",
				"f2": "my number: ************;",
				"f3": "my number: ************ ****;",
			},
		},
		{
			name: "multiple_full_replace",
			cfg: &config.Masking{
				Masks: []config.Mask{
					{
						Re:          `(\d{3})-(\d{3})-(\d{4})`,
						Mode:        config.MaskModeReplace,
						ReplaceWord: "<number>",
					},
					{
						Re:          `(test)`,
						Mode:        config.MaskModeReplace,
						ReplaceWord: "<test>",
					},
				},
			},
			input: map[string]string{
				"f1": "my number: 123_456_7890 test;",
				"f2": "my number: 123-456-7890;",
				"f3": "my number: 123-456-7890 test;",
			},
			want: map[string]string{
				"f1": "my number: 123_456_7890 <test>;",
				"f2": "my number: <number>;",
				"f3": "my number: <number> <test>;",
			},
		},
		{
			name: "multiple_full_cut",
			cfg: &config.Masking{
				Masks: []config.Mask{
					{
						Re:   `(\d{3})-(\d{3})-(\d{4})`,
						Mode: config.MaskModeCut,
					},
					{
						Re:   `(test)`,
						Mode: config.MaskModeCut,
					},
				},
			},
			input: map[string]string{
				"f1": "my number: 123_456_7890 test;",
				"f2": "my number: 123-456-7890;",
				"f3": "my number: 123-456-7890 test;",
			},
			want: map[string]string{
				"f1": "my number: 123_456_7890 ;",
				"f2": "my number: ;",
				"f3": "my number:  ;",
			},
		},
		{
			name: "single_twice_full_mask",
			cfg: &config.Masking{
				Masks: []config.Mask{
					{
						Re:   `(\d{3})-(\d{3})-(\d{4})`,
						Mode: config.MaskModeMask,
					},
				},
			},
			input: map[string]string{
				"f1": "my number: 123_456_7890 test;",
				"f2": "my number: 123-456-7890 123-456-7890;",
			},
			want: map[string]string{
				"f1": "my number: 123_456_7890 test;",
				"f2": "my number: ************ ************;",
			},
		},
		{
			name: "single_twice_full_replace",
			cfg: &config.Masking{
				Masks: []config.Mask{
					{
						Re:          `(\d{3})-(\d{3})-(\d{4})`,
						Mode:        config.MaskModeReplace,
						ReplaceWord: "<number>",
					},
				},
			},
			input: map[string]string{
				"f1": "my number: 123_456_7890 test;",
				"f2": "my number: 123-456-7890 123-456-7890;",
			},
			want: map[string]string{
				"f1": "my number: 123_456_7890 test;",
				"f2": "my number: <number> <number>;",
			},
		},
		{
			name: "single_twice_full_cut",
			cfg: &config.Masking{
				Masks: []config.Mask{
					{
						Re:   `(\d{3})-(\d{3})-(\d{4})`,
						Mode: config.MaskModeCut,
					},
				},
			},
			input: map[string]string{
				"f1": "my number: 123_456_7890 test;",
				"f2": "my number: 123-456-7890 123-456-7890;",
			},
			want: map[string]string{
				"f1": "my number: 123_456_7890 test;",
				"f2": "my number:  ;",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m, err := New(tt.cfg)
			require.NoError(t, err)

			in := tt.input
			m.Mask(in)

			for k, v := range tt.want {
				gotV, ok := in[k]
				require.True(t, ok)
				assert.Equal(t, v, gotV, "wrong value with key %q", k)
			}
		})
	}
}
