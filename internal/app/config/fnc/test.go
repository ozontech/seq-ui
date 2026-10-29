package fnc

import "errors"

// func for testing
type test struct {
	want    string
	wantErr bool
}

func parseTest(args []string) (*test, error) {
	if len(args) != 2 {
		return nil, errors.New("invalid args count")
	}
	return &test{
		want:    args[0],
		wantErr: args[1] == "true",
	}, nil
}

func (t *test) run() (string, error) {
	if t.wantErr {
		return t.want, errors.New("some err")
	}
	return t.want, nil
}
