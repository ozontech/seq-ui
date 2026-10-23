package fnc

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"

	"go.uber.org/zap"

	"github.com/ozontech/seq-ui/logger"
)

// ApplyFuncs applies funcs to config
func ApplyFuncs(v any) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return
	}
	applyFuncsReflect(rv)
}

func applyFuncsReflect(rv reflect.Value) {
	if rv.Kind() == reflect.Interface {
		if !rv.IsNil() {
			applyFuncsReflect(rv.Elem())
		}
		return
	}

	if rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return
		}
		rv = rv.Elem()
	}

	if rv.Kind() == reflect.String && rv.CanSet() {
		if newVal, ok := tryApply(rv.String()); ok {
			rv.SetString(newVal)
		}
		return
	}

	switch rv.Kind() {
	case reflect.Struct:
		for i := range rv.NumField() {
			applyFuncsReflect(rv.Field(i))
		}
	case reflect.Slice, reflect.Array:
		for i := range rv.Len() {
			applyFuncsReflect(rv.Index(i))
		}
	case reflect.Map:
		out := reflect.MakeMapWithSize(rv.Type(), rv.Len())
		iter := rv.MapRange()
		for iter.Next() {
			out.SetMapIndex(iter.Key(), applyFuncsReflectCopy(iter.Value()))
		}
		rv.Set(out)
	}
}

// map needs to be completely recreated from since its values are not addressable (CanSet() == false)
func applyFuncsReflectCopy(rv reflect.Value) reflect.Value {
	switch rv.Kind() {
	case reflect.String:
		if newVal, ok := tryApply(rv.String()); ok {
			return reflect.ValueOf(newVal)
		}
		return rv
	case reflect.Pointer, reflect.Interface:
		if rv.IsNil() {
			return rv
		}
		applyFuncsReflect(rv)
		return rv
	case reflect.Struct:
		out := reflect.New(rv.Type()).Elem()
		for i := range rv.NumField() {
			out.Field(i).Set(applyFuncsReflectCopy(rv.Field(i)))
		}
		return out
	case reflect.Slice, reflect.Array:
		out := reflect.MakeSlice(rv.Type(), rv.Len(), rv.Len())
		for i := range rv.Len() {
			out.Index(i).Set(applyFuncsReflectCopy(rv.Index(i)))
		}
		return out
	case reflect.Map:
		out := reflect.MakeMapWithSize(rv.Type(), rv.Len())
		iter := rv.MapRange()
		for iter.Next() {
			out.SetMapIndex(iter.Key(), applyFuncsReflectCopy(iter.Value()))
		}
		return out
	}

	return rv
}

// tryApply tries to parse and apply the func.
//
// Returns the result of the func execution and indicates whether it should be applied.
func tryApply(fn string) (string, bool) {
	f, err := parse(fn)
	if errors.Is(err, errEscaped) {
		return fn[1:], true
	}
	if err != nil {
		logger.Error("failed to parse config func",
			zap.String("func", fn),
			zap.Error(err),
		)
		return "", false
	}
	if f == nil {
		return "", false
	}

	res, err := f.run()
	if err != nil {
		logger.Error("failed to apply config func",
			zap.String("func", fn),
			zap.Error(err),
		)
		return "", false
	}
	return res, true
}

const (
	fnTest = "__test__" // only for testing

	fnEnv = "env"
)

var knownFuncs = []string{fnTest, fnEnv}

func isKnownFn(fn string) bool {
	return slices.Contains(knownFuncs, fn)
}

type fnc interface {
	run() (string, error)
}

var errEscaped = errors.New("escaped")

func parse(fn string) (fnc, error) {
	fnName := ""
	if idx := strings.Index(fn, "("); idx != -1 {
		fnName = fn[:idx]
		fn = fn[idx:]
	} else {
		return nil, nil
	}

	// escaped
	if strings.HasPrefix(fnName, `\`) && isKnownFn(fnName[1:]) {
		return nil, errEscaped
	}
	if !isKnownFn(fnName) {
		return nil, nil
	}

	if fn[len(fn)-1] != ')' {
		return nil, errors.New("missing ')'")
	}

	argsRaw := fn[1 : len(fn)-1]

	var args []string
	if argsRaw != "" {
		args = strings.Split(argsRaw, ",")
		for i := range args {
			args[i] = strings.TrimSpace(args[i])
		}
	}

	f, err := newFn(fnName, args)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %q func: %w", fnName, err)
	}

	return f, nil
}

func newFn(name string, args []string) (fnc, error) {
	switch name {
	case fnTest:
		return parseTest(args)
	case fnEnv:
		return parseEnv(args)
	default:
		panic("unreachable")
	}
}
