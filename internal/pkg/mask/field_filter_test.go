package mask

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ozontech/seq-ui/internal/app/config"
)

func TestFieldFilterSetCreate(t *testing.T) {
	tests := []struct {
		name string

		cfg   *config.FieldFilterSet
		noErr bool
	}{
		{
			name: "ok",
			cfg: &config.FieldFilterSet{
				Condition: config.FieldFilterConditionAnd,
				Filters: []config.FieldFilter{
					{
						Field:  "check_f1",
						Mode:   config.FieldFilterModeEqual,
						Values: []string{"test1", "test2"},
					},
					{
						Field:  "check_f2",
						Mode:   config.FieldFilterModePrefix,
						Values: []string{"pref1_", "pref2_"},
					},
				},
			},
			noErr: true,
		},
		{
			name:  "empty_cfg",
			cfg:   nil,
			noErr: true,
		},
		{
			name: "empty_filters",
			cfg: &config.FieldFilterSet{
				Filters: nil,
			},
		},
		{
			name: "unknown_condition",
			cfg: &config.FieldFilterSet{
				Condition: "unknown",
				Filters:   []config.FieldFilter{{}},
			},
		},
		{
			name: "not_condition_too_many_filters",
			cfg: &config.FieldFilterSet{
				Condition: config.FieldFilterConditionNot,
				Filters:   []config.FieldFilter{{}, {}},
			},
		},
		{
			name: "filter_empty_field",
			cfg: &config.FieldFilterSet{
				Condition: config.FieldFilterConditionAnd,
				Filters: []config.FieldFilter{
					{
						Field: "",
					},
				},
			},
		},
		{
			name: "filter_empty_values",
			cfg: &config.FieldFilterSet{
				Condition: config.FieldFilterConditionAnd,
				Filters: []config.FieldFilter{
					{
						Field:  "test",
						Values: nil,
					},
				},
			},
		},
		{
			name: "filter_unknown_mode",
			cfg: &config.FieldFilterSet{
				Condition: config.FieldFilterConditionAnd,
				Filters: []config.FieldFilter{
					{
						Field:  "test",
						Values: []string{"test1", "test2"},
						Mode:   "unknown",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := newFieldFilterSet(tt.cfg)
			require.Equal(t, tt.noErr, err == nil)
		})
	}
}

func TestFieldFilterMatch(t *testing.T) {
	tests := []struct {
		name string

		cfg config.FieldFilter

		input   map[string]string
		matched bool
	}{
		{
			name: "equal_matched",
			cfg: config.FieldFilter{
				Field:  "f1",
				Mode:   config.FieldFilterModeEqual,
				Values: []string{"test1", "test2"},
			},
			input: map[string]string{
				"f1": "test1",
			},
			matched: true,
		},
		{
			name: "equal_not_matched",
			cfg: config.FieldFilter{
				Field:  "f1",
				Mode:   config.FieldFilterModeEqual,
				Values: []string{"test1", "test2"},
			},
			input: map[string]string{
				"f1": "test3",
				"f2": "test1",
			},
			matched: false,
		},
		{
			name: "contains_matched",
			cfg: config.FieldFilter{
				Field:  "f1",
				Mode:   config.FieldFilterModeContains,
				Values: []string{"test1", "test2"},
			},
			input: map[string]string{
				"f1": "sometest1here",
			},
			matched: true,
		},
		{
			name: "contains_not_matched",
			cfg: config.FieldFilter{
				Field:  "f1",
				Mode:   config.FieldFilterModeContains,
				Values: []string{"test1", "test2"},
			},
			input: map[string]string{
				"f1": "sometest3here",
				"f2": "some test1 here",
				"f3": "sometest2here",
			},
			matched: false,
		},
		{
			name: "prefix_matched",
			cfg: config.FieldFilter{
				Field:  "f1",
				Mode:   config.FieldFilterModePrefix,
				Values: []string{"pref1_", "pref2_"},
			},
			input: map[string]string{
				"f1": "pref2_here",
			},
			matched: true,
		},
		{
			name: "prefix_not_matched",
			cfg: config.FieldFilter{
				Field:  "f1",
				Mode:   config.FieldFilterModePrefix,
				Values: []string{"pref1_", "pref2_"},
			},
			input: map[string]string{
				"f1": "pref1here",
				"f2": "pref2_here",
				"f3": "pref1_here",
			},
			matched: false,
		},
		{
			name: "suffix_matched",
			cfg: config.FieldFilter{
				Field:  "f1",
				Mode:   config.FieldFilterModeSuffix,
				Values: []string{"_suff1", "_suff2"},
			},
			input: map[string]string{
				"f1": "here_suff1",
			},
			matched: true,
		},
		{
			name: "suffix_not_matched",
			cfg: config.FieldFilter{
				Field:  "f1",
				Mode:   config.FieldFilterModeSuffix,
				Values: []string{"_suff1", "_suff2"},
			},
			input: map[string]string{
				"f1": "heresuff1",
				"f2": "here_suff2",
				"f3": "here_suff1",
			},
			matched: false,
		},
		{
			name: "field_not_found",
			cfg: config.FieldFilter{
				Field:  "f2",
				Mode:   config.FieldFilterModeEqual,
				Values: []string{"test1", "test2"},
			},
			input: map[string]string{
				"f1": "test1",
			},
			matched: false,
		},
		{
			name: "quoted",
			cfg: config.FieldFilter{
				Field:  "f1",
				Mode:   config.FieldFilterModeEqual,
				Values: []string{"test1", "test2"},
			},
			input: map[string]string{
				"f1": `"test1"`,
			},
			matched: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ff, err := newFieldFilter(tt.cfg)
			require.NoError(t, err)

			require.Equal(t, tt.matched, ff.match(tt.input))
		})
	}
}
