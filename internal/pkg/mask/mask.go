package mask

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/ozontech/seq-ui/internal/app/config"
)

const maskSymbol = byte('*')

type maskMode int8

const (
	maskModeUnknown maskMode = iota
	maskModeMask
	maskModeReplace
	maskModeCut
)

type mask struct {
	re          *regexp.Regexp
	groups      []int
	mode        maskMode
	fields      *maskFields
	replaceWord string

	fieldFilters *fieldFilterSet
}

func compileMasks(cfg []config.Mask, globalFields *maskFields) ([]mask, error) {
	m := make([]mask, 0, len(cfg))
	for i := range cfg {
		mask, err := compileMask(cfg[i], globalFields)
		if err != nil {
			return nil, fmt.Errorf("failed to compile mask #%d: %w", i, err)
		}
		m = append(m, mask)
	}
	return m, nil
}

func compileMask(cfg config.Mask, globalFields *maskFields) (mask, error) {
	if cfg.Re == "" {
		return mask{}, errors.New("empty re")
	}

	var mode maskMode
	switch cfg.Mode {
	case config.MaskModeMask:
		mode = maskModeMask
	case config.MaskModeReplace:
		mode = maskModeReplace
	case config.MaskModeCut:
		mode = maskModeCut
	default:
		return mask{}, fmt.Errorf("unknown mask mode %q", cfg.Mode)
	}

	if mode == maskModeReplace && cfg.ReplaceWord == "" {
		return mask{}, errors.New("empty replace word")
	}

	fields, err := parseFields(cfg.ProcessFields, cfg.IgnoreFields)
	if err != nil {
		return mask{}, fmt.Errorf("failed to parse fields: %w", err)
	}
	if fields == nil {
		fields = globalFields
	}

	re, err := regexp.Compile(cfg.Re)
	if err != nil {
		return mask{}, fmt.Errorf("failed to compile regexp: %w", err)
	}

	var groups []int
	if groups, err = verifyGroups(cfg.Groups, re.NumSubexp()); err != nil {
		return mask{}, fmt.Errorf("failed to verify groups: %w", err)
	}

	ff, err := newFieldFilterSet(cfg.FieldFilters)
	if err != nil {
		return mask{}, fmt.Errorf("failed to init field filters: %w", err)
	}

	return mask{
		re:           re,
		groups:       groups,
		mode:         mode,
		fields:       fields,
		replaceWord:  cfg.ReplaceWord,
		fieldFilters: ff,
	}, nil
}

func verifyGroups(groups []int, compiledTotal int) ([]int, error) {
	if len(groups) == 0 || compiledTotal == 0 || slices.Index(groups, 0) != -1 {
		return []int{0}, nil
	}

	if len(groups) > compiledTotal {
		return nil, errors.New("too many groups")
	}

	uniq := make(map[int]struct{})
	for _, g := range groups {
		if g < 0 || g > compiledTotal {
			return nil, fmt.Errorf("wrong group number %d", g)
		}
		if _, has := uniq[g]; has {
			return nil, errors.New("group numbers must be unique")
		}
		uniq[g] = struct{}{}
	}

	return groups, nil
}

// processFields returns list of fields that must be processed and
// their presence in the config
func (m *mask) processFields(event map[string]string) ([]string, bool) {
	if m.fields == nil {
		return nil, false
	}

	fields := make([]string, 0)
	if m.fields.mode == fieldsModeProcess {
		for f := range m.fields.f {
			if _, has := event[f]; has {
				fields = append(fields, f)
			}
		}
	} else {
		for f := range event {
			if _, has := m.fields.f[f]; !has {
				fields = append(fields, f)
			}
		}
	}
	return fields, true
}

func (m *mask) apply(event map[string]string) {
	if m.fieldFilters != nil && !m.fieldFilters.match(event) {
		return
	}

	fields, exists := m.processFields(event)

	if len(fields) == 0 {
		// empty list when fields presented in config
		if exists {
			return
		}

		for f, v := range event {
			event[f] = m.maskValue(v)
		}
	} else {
		for _, f := range fields {
			event[f] = m.maskValue(event[f])
		}
	}
}

func (m *mask) applyAgg(field string, bucketKeys []string) []string {
	if m.fields != nil {
		_, has := m.fields.f[field]
		if m.fields.mode == fieldsModeProcess && !has ||
			m.fields.mode == fieldsModeIgnore && has {
			return bucketKeys
		}
	}

	for i := range bucketKeys {
		bucketKeys[i] = m.maskValue(bucketKeys[i])
	}
	return bucketKeys
}

func (m *mask) maskValue(val string) string {
	if val == "" {
		return ""
	}

	indexes := m.re.FindAllStringSubmatchIndex(val, -1)
	if len(indexes) == 0 {
		return val
	}

	var sb strings.Builder
	prevFinish := 0
	curStart, curFinish := 0, 0
	for _, idx := range indexes {
		for _, grp := range m.groups {
			curStart = idx[grp*2]
			curFinish = idx[grp*2+1]
			if curStart < 0 || curFinish < 0 {
				continue
			}

			sb.WriteString(val[prevFinish:curStart])
			prevFinish = curFinish

			switch m.mode {
			case maskModeMask:
				count := curFinish - curStart
				for range count {
					sb.WriteByte(maskSymbol)
				}
			case maskModeReplace:
				sb.WriteString(m.replaceWord)
			}
		}
	}
	sb.WriteString(val[prevFinish:])

	return sb.String()
}
