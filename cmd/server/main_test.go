package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"sync"
	"syscall"
	"testing"
	"time"
)

type stubServer struct {
	startErr  error
	stopErr   error
	started   bool
	stopped   bool
	gate      chan struct{}
	closeOnce sync.Once
}

func newStubServer() *stubServer {
	return &stubServer{gate: make(chan struct{})}
}

func (s *stubServer) Start() error {
	s.started = true
	if s.startErr != nil {
		return s.startErr
	}
	<-s.gate
	return http.ErrServerClosed
}

func (s *stubServer) Stop(ctx context.Context) error {
	s.stopped = true
	s.closeOnce.Do(func() {
		close(s.gate)
	})
	return s.stopErr
}

func TestRunGracefulShutdown(t *testing.T) {
	stub := newStubServer()
	notifier := func(ch chan<- os.Signal) {
		go func() {
			ch <- syscall.SIGINT
		}()
	}

	if err := run(func() applicationServer { return stub }, notifier, 10*time.Millisecond); err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	if !stub.started || !stub.stopped {
		t.Fatalf("expected server to start and stop, got %+v", stub)
	}
}

func TestRunPropagatesStartError(t *testing.T) {
	stub := newStubServer()
	stub.startErr = errors.New("boom")

	err := run(func() applicationServer { return stub }, func(chan<- os.Signal) {}, time.Millisecond)
	if !errors.Is(err, stub.startErr) {
		t.Fatalf("expected start error, got %v", err)
	}

	if stub.stopped {
		t.Fatalf("stop should not be called when start fails")
	}
}

func TestRunPropagatesStopError(t *testing.T) {
	stub := newStubServer()
	stub.stopErr = errors.New("shutdown failure")

	notifier := func(ch chan<- os.Signal) {
		go func() {
			ch <- syscall.SIGTERM
		}()
	}

	err := run(func() applicationServer { return stub }, notifier, time.Millisecond)
	if !errors.Is(err, stub.stopErr) {
		t.Fatalf("expected stop error, got %v", err)
	}
}
