package email_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/example/iboz/internal/email"
	"github.com/example/iboz/internal/email/adapter/memory"
	"github.com/example/iboz/internal/email/adapter/synthetic"
)

type fixedClock struct {
	now time.Time
}

func (f fixedClock) Now() time.Time {
	return f.now
}

func newTestService(t *testing.T) email.ProviderService {
	t.Helper()
	repo := memory.NewRepository()
	clock := fixedClock{now: time.Date(2025, time.March, 18, 15, 30, 0, 0, time.UTC)}
	return email.NewService(repo, email.NewSHA256Hasher(), synthetic.NewGenerator(), clock)
}

func TestConfigureProviderValidation(t *testing.T) {
	svc := newTestService(t)

	ctx := context.Background()

	err := svc.ConfigureProvider(ctx, email.ProviderConfig{})
	if err == nil {
		t.Fatalf("expected error for empty config")
	}

	err = svc.ConfigureProvider(ctx, email.ProviderConfig{
		Provider:    email.ProviderIMAP,
		DisplayName: "", // missing display name
		Connection: email.ConnectionSettings{
			Protocol: email.ProtocolIMAP,
			Host:     "imap.example.com",
			Port:     993,
			UseTLS:   true,
		},
	})
	if err == nil {
		t.Fatalf("expected error for missing display name")
	}

	err = svc.ConfigureProvider(ctx, email.ProviderConfig{
		Provider:    email.ProviderIMAP,
		DisplayName: "Ops Mail",
		Connection:  email.ConnectionSettings{Protocol: email.ProtocolIMAP},
	})
	if err == nil {
		t.Fatalf("expected error for missing host/port")
	}

	err = svc.ConfigureProvider(ctx, email.ProviderConfig{
		Provider:    email.ProviderGmail,
		DisplayName: "Team Gmail",
		Connection:  email.ConnectionSettings{Protocol: email.ProtocolAPI},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAuthenticateAndFetchLifecycle(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	cfg := email.ProviderConfig{
		Provider:    email.ProviderIMAP,
		DisplayName: "Ops Mail",
		Connection: email.ConnectionSettings{
			Protocol: email.ProtocolIMAP,
			Host:     "imap.ops.local",
			Port:     993,
			UseTLS:   true,
		},
		SyncWindowHours: 48,
		LabelFilters:    []string{"Urgent", "Vendors"},
	}

	if err := svc.ConfigureProvider(ctx, cfg); err != nil {
		t.Fatalf("configure provider: %v", err)
	}

	if _, err := svc.FetchEmails(ctx); !errors.Is(err, email.ErrProviderNotAuthenticated) {
		t.Fatalf("expected auth error, got %v", err)
	}

	authState, err := svc.Authenticate(ctx, email.AuthRequest{
		Method:   email.AuthMethodAppPassword,
		Username: "ops@example.com",
		Secret:   "supersecure",
	})
	if err != nil {
		t.Fatalf("authenticate: %v", err)
	}

	if authState.Status != "connected" {
		t.Fatalf("unexpected auth status: %s", authState.Status)
	}

	messages, err := svc.FetchEmails(ctx)
	if err != nil {
		t.Fatalf("fetch emails: %v", err)
	}

	if len(messages) == 0 {
		t.Fatalf("expected synthesized messages")
	}

	state, err := svc.State(ctx)
	if err != nil {
		t.Fatalf("state: %v", err)
	}

	if state.Config == nil || state.Auth == nil {
		t.Fatalf("state missing config or auth")
	}
	if state.LastSync.IsZero() {
		t.Fatalf("expected lastSync to be set")
	}
	if state.MessagesFetched != len(messages) {
		t.Fatalf("message count mismatch: %d vs %d", state.MessagesFetched, len(messages))
	}
}

func TestAuthenticateValidation(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	if _, err := svc.Authenticate(ctx, email.AuthRequest{}); !errors.Is(err, email.ErrProviderNotConfigured) {
		t.Fatalf("expected provider not configured error, got %v", err)
	}

	if err := svc.ConfigureProvider(ctx, email.ProviderConfig{
		Provider:    email.ProviderGmail,
		DisplayName: "Ops",
		Connection:  email.ConnectionSettings{Protocol: email.ProtocolAPI},
	}); err != nil {
		t.Fatalf("configure provider: %v", err)
	}

	if _, err := svc.Authenticate(ctx, email.AuthRequest{Method: email.AuthMethodOAuth, Username: "", Secret: "abcdefghi"}); err == nil {
		t.Fatalf("expected validation error for username")
	}

	if _, err := svc.Authenticate(ctx, email.AuthRequest{Method: email.AuthMethodOAuth, Username: "ops@example.com", Secret: "short"}); err == nil {
		t.Fatalf("expected validation error for secret length")
	}
}
