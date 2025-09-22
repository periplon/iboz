package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/example/iboz/internal/server"
)

type applicationServer interface {
	Start() error
	Stop(context.Context) error
}

var (
	serverFactory  = func() applicationServer { return server.New() }
	signalNotifier = func(ch chan<- os.Signal) {
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	}
	shutdownTimeout = 10 * time.Second
)

func main() {
	if err := run(serverFactory, signalNotifier, shutdownTimeout); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func run(factory func() applicationServer, notifier func(chan<- os.Signal), timeout time.Duration) error {
	srv := factory()

	errCh := make(chan error, 1)
	go func() {
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	quit := make(chan os.Signal, 1)
	notifier(quit)

	select {
	case err, ok := <-errCh:
		if ok && err != nil {
			return err
		}
	case <-quit:
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := srv.Stop(ctx); err != nil {
		return err
	}

	return nil
}
