package grpcutil

import (
	"errors"

	"github.com/ozontech/seq-ui/internal/app/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ProcessError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, types.ErrUnauthenticated):
		return status.Error(codes.Unauthenticated, types.ErrUnauthenticated.Error())
	case errors.Is(err, types.ErrEmptyUpdateRequest) || errors.Is(err, types.ErrInvalidRequestField):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, types.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, types.ErrPermissionDenied):
		return status.Error(codes.PermissionDenied, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
