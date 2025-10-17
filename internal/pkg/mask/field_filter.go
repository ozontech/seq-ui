package mask

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/ozontech/seq-ui/internal/app/config"
)

type filterCondition int8

const (
	filterCondUnknown filterCondition = iota
	filterCondAnd
	filterCondOr
	filterCondNot
)

type fieldFilterSet struct {
	cond    filterCondition
	filters []fieldFilter
}

func newFieldFilterSet(cfg *config.FieldFilterSet) (*fieldFilterSet, error) {
	if cfg == nil {
		return nil, nil
	}

	if len(cfg.Filters) == 0 {
		return nil, errors.New("empty field filters")
	}

	var cond filterCondition
	switch cfg.Condition {
	case config.FieldFilterConditionAnd:
		cond = filterCondAnd
	case config.FieldFilterConditionOr:
		cond = filterCondOr
	case config.FieldFilterConditionNot:
		cond = filterCondNot
	default:
		return nil, fmt.Errorf("unknown field filters condition %q", cfg.Condition)
	}

	if cond == filterCondNot && len(cfg.Filters) != 1 {
		return nil, errors.New("too many filters for 'not' condition")
	}

	set := &fieldFilterSet{
		cond:    cond,
		filters: make([]fieldFilter, 0, len(cfg.Filters)),
	}

	for i := range cfg.Filters {
		ff, err := newFieldFilter(cfg.Filters[i])
		if err != nil {
			return set, fmt.Errorf("failed to init field filter #%d: %w", i, err)
		}
		set.filters = append(set.filters, ff)
	}

	return set, nil
}

func (f *fieldFilterSet) match(event map[string]string) bool {
	if f.cond == filterCondNot {
		return !f.filters[0].match(event)
	}

	for _, ff := range f.filters {
		match := ff.match(event)
		if match && f.cond == filterCondOr {
			return true
		}
		if !match && f.cond == filterCondAnd {
			return false
		}
	}

	// for 'and' - all matched, for 'or' - none
	return f.cond == filterCondAnd
}

type filterMode int8

const (
	filterModeUnknown filterMode = iota
	filterModeEqual
	filterModeContains
	filterModePrefix
	filterModeSuffix
)

type fieldFilter struct {
	mode   filterMode
	field  string
	values []string
}

func newFieldFilter(cfg config.FieldFilter) (fieldFilter, error) {
	if cfg.Field == "" {
		return fieldFilter{}, errors.New("empty field")
	}

	if len(cfg.Values) == 0 {
		return fieldFilter{}, errors.New("empty values")
	}

	var mode filterMode
	switch cfg.Mode {
	case config.FieldFilterModeEqual:
		mode = filterModeEqual
	case config.FieldFilterModeContains:
		mode = filterModeContains
	case config.FieldFilterModePrefix:
		mode = filterModePrefix
	case config.FieldFilterModeSuffix:
		mode = filterModeSuffix
	default:
		return fieldFilter{}, fmt.Errorf("unknown mode %q", cfg.Mode)
	}

	return fieldFilter{
		mode:   mode,
		field:  cfg.Field,
		values: cfg.Values,
	}, nil
}

func (f *fieldFilter) match(event map[string]string) bool {
	val, ok := event[f.field]
	if !ok {
		return false
	}

	// some map values are quoted strings
	if len(val) >= 2 && val[0] == '"' {
		unq, err := strconv.Unquote(val)
		if err == nil {
			val = unq
		}
	}

	var fn func(string, string) bool
	switch f.mode {
	case filterModeEqual:
		return slices.Contains(f.values, val)
	case filterModeContains:
		fn = strings.Contains
	case filterModePrefix:
		fn = strings.HasPrefix
	case filterModeSuffix:
		fn = strings.HasSuffix
	}

	for i := range f.values {
		if len(val) < len(f.values[i]) {
			continue
		}
		if fn(val, f.values[i]) {
			return true
		}
	}
	return false
}
