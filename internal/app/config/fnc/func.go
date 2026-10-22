package fnc

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"go.uber.org/zap"

	"github.com/ozontech/seq-ui/logger"
)

// TryApply tries to parse and apply the func.
//
// Returns the result of the func execution and indicates whether it should be applied.
func TryApply(fn string) (string, bool) {
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
