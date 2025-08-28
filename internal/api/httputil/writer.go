package httputil

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	headerContentType = "Content-Type"

	contentTypeJSON = "application/json"
)

type Writer struct {
	http.ResponseWriter
	ErrorMessage string
	StatusCode   int
}

func NewWriter(w http.ResponseWriter) *Writer {
	if _, ok := w.(*Writer); ok {
		return w.(*Writer)
	}
	w.Header().Set(headerContentType, contentTypeJSON)
	return &Writer{ResponseWriter: w}
}

func (w *Writer) WriteHeader(statusCode int) {
	w.StatusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *Writer) WriteJson(data any) {
	resp, err := json.Marshal(data)
	if err != nil {
		w.Error(err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(resp)
}

func (w *Writer) Error(err error, code int) {
	msg := err.Error()
	if st, ok := status.FromError(err); ok {
		msg = st.Message()
		code = grpcToHttpCode(st.Code())
	}

	resp, _ := json.Marshal(Error{
		Message: msg,
	})

	w.WriteHeader(code)
	_, _ = w.Write(resp)
	w.ErrorMessage = msg
}

func grpcToHttpCode(code codes.Code) int {
	switch code {
	case codes.OK:
		return http.StatusOK
	case codes.Canceled:
		return 499
	case codes.InvalidArgument, codes.OutOfRange, codes.FailedPrecondition:
		return http.StatusBadRequest
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists, codes.Aborted:
		return http.StatusConflict
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests
	case codes.Unimplemented:
		return http.StatusNotImplemented
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	case codes.Unknown, codes.DataLoss, codes.Internal:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

type ChunkedWriter struct {
	w http.ResponseWriter
}

func NewChunkedWriter(w http.ResponseWriter) (*ChunkedWriter, error) {
	if _, ok := w.(http.Flusher); !ok {
		return nil, errors.New("http.ResponseWriter is not http.Flusher")
	}
	w.Header().Set(headerContentType, contentTypeJSON)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	return &ChunkedWriter{w: w}, nil
}

func (w *ChunkedWriter) WriteString(data string) {
	_, _ = fmt.Fprintf(w.w, "%s\r\n", data)
}

func (w *ChunkedWriter) Write(p []byte) (int, error) {
	return w.w.Write(p)
}

func (w *ChunkedWriter) Flush() {
	w.w.(http.Flusher).Flush()
}
