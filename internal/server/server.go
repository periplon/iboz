package server

import (
	"context"
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/example/iboz/internal/api"
)

//go:embed all:static
var embeddedStatic embed.FS

const (
	defaultAddress = ":8080"
	readTimeout    = 15 * time.Second
	writeTimeout   = 15 * time.Second
)

type Server struct {
	httpServer *http.Server
}

func New() *Server {
	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.RequestID())
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	api.Register(e.Group("/api"))

	subFS, err := fs.Sub(embeddedStatic, "static")
	if err != nil {
		log.Fatalf("failed to load embedded frontend: %v", err)
	}

	e.GET("/*", spaHandler(http.FS(subFS)))

	addr := defaultAddress
	if fromEnv := os.Getenv("IBOZ_LISTEN_ADDR"); fromEnv != "" {
		addr = fromEnv
	}

	srv := &http.Server{
		Addr:         addr,
		Handler:      e,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}

	return &Server{httpServer: srv}
}

func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func spaHandler(filesystem http.FileSystem) echo.HandlerFunc {
	fileServer := http.FileServer(filesystem)

	return func(c echo.Context) error {
		requestPath := c.Request().URL.Path
		cleanedPath := strings.TrimPrefix(requestPath, "/")
		if cleanedPath == "" {
			fileServer.ServeHTTP(c.Response(), c.Request())
			return nil
		}

		f, err := filesystem.Open(cleanedPath)
		if err != nil {
			c.Request().URL.Path = "/"
			fileServer.ServeHTTP(c.Response(), c.Request())
			return nil
		}
		defer f.Close()

		info, err := f.Stat()
		if err != nil || info.IsDir() {
			c.Request().URL.Path = "/"
		}

		fileServer.ServeHTTP(c.Response(), c.Request())
		return nil
	}
}
