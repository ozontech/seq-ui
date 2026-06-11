package httputil

import (
	"bytes"
	"encoding/json"
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

type TestDataHTTPEx[Treq, Tresp any] struct {
	Method string
	Target string
	Req    Treq

	Handler http.HandlerFunc

	Want    Tresp
	NoResp  bool
	WantErr bool
}

func DoTestHTTPEx[Treq, Tresp any](t *testing.T, data TestDataHTTPEx[Treq, Tresp]) {
	w := httptest.NewRecorder()

	reqBody, err := json.Marshal(data.Req)
	require.NoError(t, err)
	req := httptest.NewRequest(data.Method, data.Target, bytes.NewReader(reqBody))

	data.Handler(w, req)

	resp := w.Result()
	defer func() {
		_ = resp.Body.Close()
	}()

	require.Equal(t, data.WantErr, resp.StatusCode != http.StatusOK)
	if data.WantErr || data.NoResp {
		return
	}

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	want, err := json.Marshal(data.Want)
	require.NoError(t, err)

	require.Equal(t, want, respBody)
}
