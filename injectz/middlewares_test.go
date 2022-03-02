package injectz_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ibrt/golang-inject/injectz"
	"github.com/ibrt/golang-inject/injectz/internal/fixtures"
)

func TestHTTPMiddleware(t *testing.T) {
	injector := injectz.NewSingletonInjector(fixtures.FirstContextKey, true)
	middleware := injectz.NewMiddleware(injector)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.True(t, r.Context().Value(fixtures.FirstContextKey).(bool))
		_, _ = w.Write([]byte("ok"))
	})

	srv := httptest.NewServer(middleware(handler))
	defer srv.Close()

	resp, err := http.Get(srv.URL)
	require.NoError(t, err)
	defer func() {
		_ = resp.Body.Close()
	}()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	buf, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "ok", string(buf))
}
