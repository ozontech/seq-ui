package httputil

import (
	"errors"
	"net/http"

	"github.com/ozontech/seq-ui/internal/app/types"
)

type Error struct {
	Message string `json:"message"`
} // @name UnexpectedError

func ProcessError(w *Writer, err error) {
	switch {
	case err == nil:
		return
	case errors.Is(err, types.ErrUnauthenticated) || errors.Is(err, types.ErrBadUserKeyValueType):
		w.Error(err, http.StatusUnauthorized)
	case errors.Is(err, types.ErrEmptyUpdateRequest) || errors.Is(err, types.ErrInvalidRequestField):
		w.Error(err, http.StatusBadRequest)
	case errors.Is(err, types.ErrNotFound):
		w.Error(err, http.StatusNotFound)
	case errors.Is(err, types.ErrPermissionDenied):
		w.Error(err, http.StatusForbidden)
	default:
		w.Error(err, http.StatusInternalServerError)
	}
}
