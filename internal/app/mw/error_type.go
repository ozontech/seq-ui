package mw

import "google.golang.org/grpc/codes"

// respErrorType unification for grpc and http status code types.
type respErrorType int

const (
	respNoError respErrorType = iota
	respClientError
	respServerError
)

// gRPCRespErrorTypeFromStatusCode returns response error type depending on gRPC status code.
func gRPCRespErrorTypeFromStatusCode(statusCode codes.Code) respErrorType {
	switch statusCode {
	case codes.OK:
		return respNoError
	case codes.InvalidArgument, codes.NotFound, codes.AlreadyExists,
		codes.PermissionDenied, codes.Unauthenticated,
		codes.FailedPrecondition, codes.OutOfRange:
		return respClientError
	default:
		return respServerError
	}
}

// httpRespErrorTypeFromStatusCode returns response error type depending on HTTP status code.
func httpRespErrorTypeFromStatusCode(statusCode int) respErrorType {
	switch {
	case statusCode < 400:
		return respNoError
	case statusCode < 500:
		return respClientError
	default:
		return respServerError
	}
}
