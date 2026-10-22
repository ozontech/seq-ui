package fnc

import (
	"errors"
	"os"
)

type env struct {
	args []string
}

func parseEnv(args []string) (*env, error) {
	if len(args) != 1 {
		return nil, errors.New("invalid args count")
	}
	if args[0] == "" {
		return nil, errors.New("empty arg")
	}
	return &env{args: args}, nil
}

func (e *env) run() (string, error) {
	env, ok := os.LookupEnv(e.args[0])
	if !ok {
		return "", errors.New("failed to lookup env")
	}
	return env, nil
}
