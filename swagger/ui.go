package swagger

import (
	"net/http"

	statik "github.com/rakyll/statik/fs"
)

const swaggerUI = "swagger"

// GetUI returns a file system in which the Swagger UI is embedded.
func GetUI() http.FileSystem {
	statik.RegisterWithNamespace(swaggerUI, swaggerUIZip)

	f, err := statik.NewWithNamespace(swaggerUI)
	if err != nil {
		panic(err)
	}

	return f
}
