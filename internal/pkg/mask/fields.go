package mask

import "errors"

type fieldsMode int8

const (
	fieldsModeUnknown fieldsMode = iota
	fieldsModeProcess
	fieldsModeIgnore
)

type maskFields struct {
	f    map[string]struct{}
	mode fieldsMode
}

func parseFields(process, ignore []string) (*maskFields, error) {
	newFields := func(fields []string, mode fieldsMode) *maskFields {
		m := map[string]struct{}{}
		for _, f := range fields {
			m[f] = struct{}{}
		}
		return &maskFields{
			f:    m,
			mode: mode,
		}
	}

	switch {
	case len(process) > 0 && len(ignore) > 0:
		return nil, errors.New("igore and process fields cannot be specified at the same time")
	case len(process) == 0 && len(ignore) == 0:
		return nil, nil
	case len(process) > 0:
		return newFields(process, fieldsModeProcess), nil
	default:
		return newFields(ignore, fieldsModeIgnore), nil
	}
}
