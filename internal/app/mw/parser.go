package mw

import (
	"errors"
	"fmt"
	"strings"
)

// parseURI parses URI into api, version and root method.
//
// Example 1: /api/v1/method/other -> (api, v1, method)
//
// Example 2: /api/v2/method -> (api, v2, method)
func parseURI(uri string) (string, string, string, error) {
	if uri == "" || uri == "/" || uri[0] != '/' {
		return "", "", "", fmt.Errorf("incorrect URI format: %q", uri)
	}

	uriParts := strings.Split(uri[1:], "/")
	if len(uriParts) < 3 {
		return "", "", "", fmt.Errorf("not enough parts of URI: %q", uri)
	}

	return uriParts[0], uriParts[1], uriParts[2], nil
}

// parseGRPCFullMethod parses FullMethod into service and method pair.
//
// Example 1: /api.v1.Service/Method -> (Service, Method)
//
// Example 2: /Service/Method -> (Service, Method)
func parseGRPCFullMethod(fullMethod string) (string, string, error) {
	if fullMethod == "" {
		return "", "", errors.New("empty FullMethod")
	}
	if fullMethod[0] != '/' {
		return "", "", fmt.Errorf("FullMethod must start from slash '/': %q", fullMethod)
	}

	// skip first slash
	sliceStrings := strings.Split(fullMethod[1:], "/")
	if len(sliceStrings) != 2 {
		return "", "", fmt.Errorf("incorrect FullMethod format: %q", fullMethod)
	}

	// service name is after last dot
	lastDotIndex := strings.LastIndex(sliceStrings[0], ".")

	return sliceStrings[0][lastDotIndex+1:], sliceStrings[1], nil
}
