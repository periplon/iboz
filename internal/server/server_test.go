package server

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/labstack/echo/v4"
)

func TestNewConfiguresHTTPServer(t *testing.T) {
	t.Setenv("IBOZ_LISTEN_ADDR", "")

	srv := New()
	if srv.httpServer == nil {
		t.Fatalf("httpServer was nil")
	}

	if srv.httpServer.Addr != defaultAddress {
		t.Fatalf("expected default address %q, got %q", defaultAddress, srv.httpServer.Addr)
	}

	if srv.httpServer.Handler == nil {
		t.Fatalf("http server handler was nil")
	}

	if srv.httpServer.ReadTimeout != readTimeout {
		t.Fatalf("unexpected read timeout: %s", srv.httpServer.ReadTimeout)
	}

	if srv.httpServer.WriteTimeout != writeTimeout {
		t.Fatalf("unexpected write timeout: %s", srv.httpServer.WriteTimeout)
	}
}

func TestNewHonorsEnvironmentAddress(t *testing.T) {
	t.Setenv("IBOZ_LISTEN_ADDR", ":9090")

	srv := New()
	if srv.httpServer.Addr != ":9090" {
		t.Fatalf("expected address to honor env var, got %q", srv.httpServer.Addr)
	}
}

func setupSPAHandler(t *testing.T) echo.HandlerFunc {
	t.Helper()
	fsys := fstest.MapFS{
		"index.html": {Data: []byte("index")},
		"app.js":     {Data: []byte("console.log('hi')")},
		"dir":        {Mode: fs.ModeDir},
	}

	handler := spaHandler(http.FS(fsys))
	return handler
}

func executeSPARequest(t *testing.T, handler echo.HandlerFunc, target string) *httptest.ResponseRecorder {
	t.Helper()
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, target, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := handler(c); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	return rec
}

func TestSpaHandlerServesRootIndex(t *testing.T) {
	handler := setupSPAHandler(t)

	rec := executeSPARequest(t, handler, "/")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	if body := rec.Body.String(); body != "index" {
		t.Fatalf("unexpected body: %q", body)
	}
}

func TestSpaHandlerServesExistingFile(t *testing.T) {
	handler := setupSPAHandler(t)

	rec := executeSPARequest(t, handler, "/app.js")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	if body := rec.Body.String(); body != "console.log('hi')" {
		t.Fatalf("unexpected body: %q", body)
	}
}

func TestSpaHandlerFallsBackForMissingPath(t *testing.T) {
	handler := setupSPAHandler(t)

	rec := executeSPARequest(t, handler, "/missing")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	if body := rec.Body.String(); body != "index" {
		t.Fatalf("expected index fallback, got %q", body)
	}
}

func TestSpaHandlerFallsBackForDirectory(t *testing.T) {
	handler := setupSPAHandler(t)

	rec := executeSPARequest(t, handler, "/dir")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	if body := rec.Body.String(); body != "index" {
		t.Fatalf("expected index fallback for directory, got %q", body)
	}
}
