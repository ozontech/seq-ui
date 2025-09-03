package swagger

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/go-openapi/spec"
)

//go:embed swagger.json
var swaggerJSON []byte

type SpecPath spec.PathItem

func (sp SpecPath) getOperation(method string) *spec.Operation {
	switch method {
	case http.MethodGet, "":
		return sp.Get
	case http.MethodHead:
		return sp.Head
	case http.MethodPost:
		return sp.Post
	case http.MethodPut:
		return sp.Put
	case http.MethodPatch:
		return sp.Patch
	case http.MethodDelete:
		return sp.Delete
	case http.MethodOptions:
		return sp.Options
	default:
		return nil
	}
}

func (sp SpecPath) HasSecurity(method string) bool {
	op := sp.getOperation(method)
	return op != nil && len(op.Security) != 0
}

type Spec spec.Swagger

func (s Spec) FindPath(path string) (SpecPath, bool) {
	if s.Paths == nil {
		return SpecPath{}, false
	}
	p, ok := s.Paths.Paths[path]
	return SpecPath(p), ok
}

func (s Spec) HasApi(api string) bool {
	if s.Paths == nil {
		return false
	}
	for path := range s.Paths.Paths {
		if strings.HasPrefix(path, fmt.Sprintf("/%s/", api)) {
			return true
		}
	}
	return false
}

var (
	swaggerSpec spec.Swagger

	initSpec = sync.OnceFunc(func() {
		_ = json.Unmarshal(swaggerJSON, &swaggerSpec)
	})
)

// GetSpec returns OpenAPI spec of the SeqUI server.
func GetSpec() spec.Swagger {
	initSpec()
	return swaggerSpec
}

// GetSpecEx returns extended OpenAPI spec of the SeqUI server.
func GetSpecEx() Spec {
	return Spec(GetSpec())
}
