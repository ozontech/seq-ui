package httputil

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestDataHTTP struct {
	Req     *http.Request
	Handler http.HandlerFunc

	WantRespBody string
	WantStatus   int
}

func DoTestHTTP(t *testing.T, data TestDataHTTP) {
	w := httptest.NewRecorder()

	data.Handler(w, data.Req)

	resp := w.Result()
	defer func() {
		_ = resp.Body.Close()
	}()

	assert.Equal(t, data.WantStatus, resp.StatusCode)

	if data.WantRespBody != "" {
		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, data.WantRespBody, string(respBody))
	}
}
