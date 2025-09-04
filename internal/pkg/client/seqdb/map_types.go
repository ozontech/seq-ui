package seqdb

import (
	"bytes"
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

func (m mapStringJson) getValues(keys []string, raw bool) []string {
	values := make([]string, 0, len(keys))
	for _, k := range keys {
		if v, ok := m[k]; ok {
			val := marshalRaw(v)
			if !raw {
				val = bytes.TrimPrefix(val, []byte{'"'})
				val = bytes.TrimSuffix(val, []byte{'"'})
			}
			values = append(values, validateString(string(val)))
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

type mapStringString map[string]string

func newMapStringString(data []byte) (mapStringString, error) {
	m, err := newMapStringJson(data)
	if err != nil {
		return nil, err
	}
	return m.toStringMap(), nil
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
