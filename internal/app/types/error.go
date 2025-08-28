package types

import (
	"errors"
	"fmt"
)

var (
	ErrEmptyUpdateRequest    = errors.New("empty update request")
	ErrInvalidRequestField   = errors.New("invalid request field")
	ErrNotFound              = errors.New("not found")
	ErrPermissionDenied      = errors.New("permission denied")
	ErrAsyncSearchesDisabled = errors.New("async searches disabled")
)

func NewErrInvalidRequestField(err string) error {
	return fmt.Errorf("%w: %s", ErrInvalidRequestField, err)
}

func NewErrNotFound(obj string) error {
	return fmt.Errorf("%w: %s", ErrNotFound, obj)
}

func NewErrPermissionDenied(operation string) error {
	return fmt.Errorf("%w: %s", ErrPermissionDenied, operation)
}
