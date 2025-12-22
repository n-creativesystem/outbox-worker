package healthcheck

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHealthCheckServer(t *testing.T) {
	noop := NewNoop()
	srv := New(noop, "")
	server := httptest.NewServer(srv)
	u, err := url.Parse(server.URL)
	require.NoError(t, err)
	u = u.JoinPath("/healthz")
	resp, err := http.Get(u.String())
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, resp.StatusCode, 200)
}

func TestHealthCheckServerWithErr(t *testing.T) {
	noop := newNoopWithErr(errors.New("error"))
	srv := New(noop, "")
	server := httptest.NewServer(srv)
	u, err := url.Parse(server.URL)
	require.NoError(t, err)
	u = u.JoinPath("/healthz")
	resp, err := http.Get(u.String())
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, resp.StatusCode, 500)
}
