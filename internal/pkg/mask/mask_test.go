package mask

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ozontech/seq-ui/internal/app/config"
)

func TestMaskCompile(t *testing.T) {
	tests := []struct {
		name string

		cfg   config.Mask
		noErr bool
	}{
		{
			name: "ok",
			cfg: config.Mask{
				Re:            `(\d{3})-(\d{3})-(\d{4})`,
				Groups:        []int{0},
				Mode:          config.MaskModeReplace,
				ReplaceWord:   "test",
				ProcessFields: []string{"f1", "f2"},
			},
			noErr: true,
		},
		{
			name: "empty_re",
			cfg: config.Mask{
				Re: "",
			},
		},
		{
			name: "bad_re",
			cfg: config.Mask{
				Re: "(test",
			},
		},
		{
			name: "unknown_mask_mode",
			cfg: config.Mask{
				Re:   "(test)",
				Mode: "unknown",
			},
		},
		{
			name: "empty_replace_word",
			cfg: config.Mask{
				Re:          "(test)",
				Mode:        config.MaskModeReplace,
				ReplaceWord: "",
			},
		},
		{
			name: "process_and_ignore",
			cfg: config.Mask{
				Re:            "(test)",
				Mode:          config.MaskModeMask,
				ProcessFields: []string{"f1", "f2"},
				IgnoreFields:  []string{"f3", "f4"},
			},
		},
		{
			name: "process_and_ignore",
			cfg: config.Mask{
				Re:            "(test)",
				Mode:          config.MaskModeMask,
				ProcessFields: []string{"f1", "f2"},
				IgnoreFields:  []string{"f3", "f4"},
			},
		},
		{
			name: "too_many_groups",
			cfg: config.Mask{
				Re:     "(test)",
				Mode:   config.MaskModeMask,
				Groups: []int{1, 2},
			},
		},
		{
			name: "wrong_group_number_1",
			cfg: config.Mask{
				Re:     "(test)",
				Mode:   config.MaskModeMask,
				Groups: []int{-1},
			},
		},
		{
			name: "wrong_group_number_2",
			cfg: config.Mask{
				Re:     "(test)",
				Mode:   config.MaskModeMask,
				Groups: []int{10},
			},
		},
		{
			name: "group_not_unique",
			cfg: config.Mask{
				Re:     "(test)-(test2)",
				Mode:   config.MaskModeMask,
				Groups: []int{1, 1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := compileMask(tt.cfg, nil)
			require.Equal(t, tt.noErr, err == nil)
		})
	}
}

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
			name: "no_groups",
			cfg: config.Mask{
				Re:   `(\d{3})-(\d{3})-(\d{4})`,
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
		{
			name: "no_groups_in_re",
			cfg: config.Mask{
				Re:     `\d{3}-\d{3}-\d{4}`,
				Groups: []int{1, 3},
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
			name: "groups_with_zero",
			cfg: config.Mask{
				Re:     `(\d{3})-(\d{3})-(\d{4})`,
				Groups: []int{0, 1, 3},
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
			name: "process_local_not_exists",
			cfg: config.Mask{
				Re:            `(\d{3})-(\d{3})-(\d{4})`,
				Groups:        []int{0},
				Mode:          config.MaskModeMask,
				ProcessFields: []string{"f3"},
			},
			globalProcessFields: []string{"f2"},
			input: map[string]string{
				"f1": "my number: 098-765-4321;",
				"f2": "my number: 123-456-7890;",
			},
			want: map[string]string{
				"f1": "my number: 098-765-4321;",
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
			name: "process_global_not_exists",
			cfg: config.Mask{
				Re:     `(\d{3})-(\d{3})-(\d{4})`,
				Groups: []int{0},
				Mode:   config.MaskModeMask,
			},
			globalProcessFields: []string{"f3"},
			input: map[string]string{
				"f1": "my number: 098-765-4321;",
				"f2": "my number: 123-456-7890;",
			},
			want: map[string]string{
				"f1": "my number: 098-765-4321;",
				"f2": "my number: 123-456-7890;",
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
			name: "ignore_local_all",
			cfg: config.Mask{
				Re:           `(\d{3})-(\d{3})-(\d{4})`,
				Groups:       []int{0},
				Mode:         config.MaskModeMask,
				IgnoreFields: []string{"f1", "f2"},
			},
			globalIgnoreFields: []string{"f2"},
			input: map[string]string{
				"f1": "my number: 098-765-4321;",
				"f2": "my number: 123-456-7890;",
			},
			want: map[string]string{
				"f1": "my number: 098-765-4321;",
				"f2": "my number: 123-456-7890;",
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
			name: "ignore_global_all",
			cfg: config.Mask{
				Re:     `(\d{3})-(\d{3})-(\d{4})`,
				Groups: []int{0},
				Mode:   config.MaskModeMask,
			},
			globalIgnoreFields: []string{"f1", "f2"},
			input: map[string]string{
				"f1": "my number: 098-765-4321;",
				"f2": "my number: 123-456-7890;",
			},
			want: map[string]string{
				"f1": "my number: 098-765-4321;",
				"f2": "my number: 123-456-7890;",
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

func TestMaskApplyAgg(t *testing.T) {
	tests := []struct {
		name string

		cfg                 config.Mask
		globalProcessFields []string
		globalIgnoreFields  []string

		field        string
		inputBuckets []string
		wantBuckets  []string
	}{
		{
			name: "full_mask",
			cfg: config.Mask{
				Re:     `(\d{3})-(\d{3})-(\d{4})`,
				Groups: []int{0},
				Mode:   config.MaskModeMask,
			},
			inputBuckets: []string{
				"my number: 123_456_7890;",
				"my number: 123-456-7890;",
			},
			wantBuckets: []string{
				"my number: 123_456_7890;",
				"my number: ************;",
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
			inputBuckets: []string{
				"my number: 123_456_7890;",
				"my number: 123-456-7890;",
			},
			wantBuckets: []string{
				"my number: 123_456_7890;",
				"my number: <number>;",
			},
		},
		{
			name: "full_cut",
			cfg: config.Mask{
				Re:     `(\d{3})-(\d{3})-(\d{4})`,
				Groups: []int{0},
				Mode:   config.MaskModeCut,
			},
			inputBuckets: []string{
				"my number: 123_456_7890;",
				"my number: 123-456-7890;",
			},
			wantBuckets: []string{
				"my number: 123_456_7890;",
				"my number: ;",
			},
		},
		{
			name: "groups",
			cfg: config.Mask{
				Re:     `(\d{3})-(\d{3})-(\d{4})`,
				Groups: []int{1, 3},
				Mode:   config.MaskModeMask,
			},
			inputBuckets: []string{
				"my number: 123_456_7890;",
				"my number: 123-456-7890;",
			},
			wantBuckets: []string{
				"my number: 123_456_7890;",
				"my number: ***-456-****;",
			},
		},
		{
			name: "no_groups",
			cfg: config.Mask{
				Re:   `(\d{3})-(\d{3})-(\d{4})`,
				Mode: config.MaskModeMask,
			},
			inputBuckets: []string{
				"my number: 123_456_7890;",
				"my number: 123-456-7890;",
			},
			wantBuckets: []string{
				"my number: 123_456_7890;",
				"my number: ************;",
			},
		},
		{
			name: "no_groups_in_re",
			cfg: config.Mask{
				Re:     `\d{3}-\d{3}-\d{4}`,
				Groups: []int{1, 3},
				Mode:   config.MaskModeMask,
			},
			inputBuckets: []string{
				"my number: 123_456_7890;",
				"my number: 123-456-7890;",
			},
			wantBuckets: []string{
				"my number: 123_456_7890;",
				"my number: ************;",
			},
		},
		{
			name: "groups_with_zero",
			cfg: config.Mask{
				Re:     `(\d{3})-(\d{3})-(\d{4})`,
				Groups: []int{0, 1, 3},
				Mode:   config.MaskModeMask,
			},
			inputBuckets: []string{
				"my number: 123_456_7890;",
				"my number: 123-456-7890;",
			},
			wantBuckets: []string{
				"my number: 123_456_7890;",
				"my number: ************;",
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
			field:               "f1",
			inputBuckets: []string{
				"my number: 123_456_7890;",
				"my number: 123-456-7890;",
			},
			wantBuckets: []string{
				"my number: 123_456_7890;",
				"my number: ************;",
			},
		},
		{
			name: "process_local_not_exists",
			cfg: config.Mask{
				Re:            `(\d{3})-(\d{3})-(\d{4})`,
				Groups:        []int{0},
				Mode:          config.MaskModeMask,
				ProcessFields: []string{"f2"},
			},
			globalProcessFields: []string{"f1"},
			field:               "f1",
			inputBuckets: []string{
				"my number: 123_456_7890;",
				"my number: 123-456-7890;",
			},
			wantBuckets: []string{
				"my number: 123_456_7890;",
				"my number: 123-456-7890;",
			},
		},
		{
			name: "process_global",
			cfg: config.Mask{
				Re:     `(\d{3})-(\d{3})-(\d{4})`,
				Groups: []int{0},
				Mode:   config.MaskModeMask,
			},
			globalProcessFields: []string{"f1"},
			field:               "f1",
			inputBuckets: []string{
				"my number: 123_456_7890;",
				"my number: 123-456-7890;",
			},
			wantBuckets: []string{
				"my number: 123_456_7890;",
				"my number: ************;",
			},
		},
		{
			name: "process_global_not_exists",
			cfg: config.Mask{
				Re:     `(\d{3})-(\d{3})-(\d{4})`,
				Groups: []int{0},
				Mode:   config.MaskModeMask,
			},
			globalProcessFields: []string{"f2"},
			field:               "f1",
			inputBuckets: []string{
				"my number: 123_456_7890;",
				"my number: 123-456-7890;",
			},
			wantBuckets: []string{
				"my number: 123_456_7890;",
				"my number: 123-456-7890;",
			},
		},
		{
			name: "ignore_local_masked",
			cfg: config.Mask{
				Re:           `(\d{3})-(\d{3})-(\d{4})`,
				Groups:       []int{0},
				Mode:         config.MaskModeMask,
				IgnoreFields: []string{"f2"},
			},
			globalIgnoreFields: []string{"f1"},
			field:              "f1",
			inputBuckets: []string{
				"my number: 123_456_7890;",
				"my number: 123-456-7890;",
			},
			wantBuckets: []string{
				"my number: 123_456_7890;",
				"my number: ************;",
			},
		},
		{
			name: "ignore_local_not_masked",
			cfg: config.Mask{
				Re:           `(\d{3})-(\d{3})-(\d{4})`,
				Groups:       []int{0},
				Mode:         config.MaskModeMask,
				IgnoreFields: []string{"f1"},
			},
			globalIgnoreFields: []string{"f2"},
			field:              "f1",
			inputBuckets: []string{
				"my number: 123_456_7890;",
				"my number: 123-456-7890;",
			},
			wantBuckets: []string{
				"my number: 123_456_7890;",
				"my number: 123-456-7890;",
			},
		},
		{
			name: "ignore_global_masked",
			cfg: config.Mask{
				Re:     `(\d{3})-(\d{3})-(\d{4})`,
				Groups: []int{0},
				Mode:   config.MaskModeMask,
			},
			globalIgnoreFields: []string{"f2"},
			field:              "f1",
			inputBuckets: []string{
				"my number: 123_456_7890;",
				"my number: 123-456-7890;",
			},
			wantBuckets: []string{
				"my number: 123_456_7890;",
				"my number: ************;",
			},
		},
		{
			name: "ignore_global_not_masked",
			cfg: config.Mask{
				Re:     `(\d{3})-(\d{3})-(\d{4})`,
				Groups: []int{0},
				Mode:   config.MaskModeMask,
			},
			globalIgnoreFields: []string{"f1"},
			field:              "f1",
			inputBuckets: []string{
				"my number: 123_456_7890;",
				"my number: 123-456-7890;",
			},
			wantBuckets: []string{
				"my number: 123_456_7890;",
				"my number: 123-456-7890;",
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

			inBuckets := tt.inputBuckets
			inBuckets = mask.applyAgg(tt.field, inBuckets)

			for i, v := range tt.wantBuckets {
				gotV := inBuckets[i]
				assert.Equal(t, v, gotV, "wrong bucket with index %d", i)
			}
		})
	}
}
