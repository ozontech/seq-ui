package seqdb

import (
	"strconv"
	"strings"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type mapStringJson map[string]jsoniter.RawMessage

func newMapStringJson(data []byte) (mapStringJson, error) {
	m := make(mapStringJson)
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return m, nil
}

func (m mapStringJson) toStringMap() mapStringString {
	mss := make(mapStringString, len(m))
	for k, v := range m {
		mss[k] = validateString(string(marshalRaw(v)))
	}
	return mss
}

type mapStringString map[string]string

func newMapStringString(data []byte) (mapStringString, error) {
	m, err := newMapStringJson(data)
	if err != nil {
		return nil, err
	}
	return m.toStringMap(), nil
}

func (m mapStringString) getValues(keys []string, raw bool) []string {
	values := make([]string, 0, len(keys))
	for _, k := range keys {
		if v, ok := m[k]; ok {
			if !raw {
				if unq, err := strconv.Unquote(v); err == nil {
					v = unq
				}
			}
			values = append(values, v)
		} else {
			val := ""
			if raw {
				val = `""`
			}
			values = append(values, val)
		}
	}
	return values
}

const invalidUTF8Replacement = "�"

func validateString(s string) string {
	return strings.ToValidUTF8(s, invalidUTF8Replacement)
}

func marshalRaw(raw jsoniter.RawMessage) []byte {
	if raw == nil {
		return []byte("null")
	}
	return raw
}
