package email

import (
	"context"
	"errors"
	"testing"
)

func TestConfigureProviderValidation(t *testing.T) {
	svc := NewService()

	err := svc.ConfigureProvider(ProviderConfig{})
	if err == nil {
		t.Fatalf("expected error for empty config")
	}

	err = svc.ConfigureProvider(ProviderConfig{
		Provider:    ProviderIMAP,
		DisplayName: "", // missing display name
		Connection: ConnectionSettings{
			Protocol: ProtocolIMAP,
			Host:     "imap.example.com",
			Port:     993,
			UseTLS:   true,
		},
	})
	if err == nil {
		t.Fatalf("expected error for missing display name")
	}

	err = svc.ConfigureProvider(ProviderConfig{
		Provider:    ProviderIMAP,
		DisplayName: "Ops Mail",
		Connection:  ConnectionSettings{Protocol: ProtocolIMAP},
	})
	if err == nil {
		t.Fatalf("expected error for missing host/port")
	}

	err = svc.ConfigureProvider(ProviderConfig{
		Provider:    ProviderGmail,
		DisplayName: "Team Gmail",
		Connection:  ConnectionSettings{Protocol: ProtocolAPI},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAuthenticateAndFetchLifecycle(t *testing.T) {
	svc := NewService()
	cfg := ProviderConfig{
		Provider:    ProviderIMAP,
		DisplayName: "Ops Mail",
		Connection: ConnectionSettings{
			Protocol: ProtocolIMAP,
			Host:     "imap.ops.local",
			Port:     993,
			UseTLS:   true,
		},
		SyncWindowHours: 48,
		LabelFilters:    []string{"Urgent", "Vendors"},
	}

	if err := svc.ConfigureProvider(cfg); err != nil {
		t.Fatalf("configure provider: %v", err)
	}

	if _, err := svc.FetchEmails(context.Background()); !errors.Is(err, ErrProviderNotAuthenticated) {
		t.Fatalf("expected auth error, got %v", err)
	}

	authState, err := svc.Authenticate(AuthRequest{
		Method:   AuthMethodAppPassword,
		Username: "ops@example.com",
		Secret:   "supersecure",
	})
	if err != nil {
		t.Fatalf("authenticate: %v", err)
	}

	if authState.Status != "connected" {
		t.Fatalf("unexpected auth status: %s", authState.Status)
	}

	messages, err := svc.FetchEmails(context.Background())
	if err != nil {
		t.Fatalf("fetch emails: %v", err)
	}

	if len(messages) == 0 {
		t.Fatalf("expected synthesized messages")
	}

	state := svc.State()
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
	svc := NewService()

	if _, err := svc.Authenticate(AuthRequest{}); !errors.Is(err, ErrProviderNotConfigured) {
		t.Fatalf("expected provider not configured error, got %v", err)
	}

	if err := svc.ConfigureProvider(ProviderConfig{
		Provider:    ProviderGmail,
		DisplayName: "Ops",
		Connection:  ConnectionSettings{Protocol: ProtocolAPI},
	}); err != nil {
		t.Fatalf("configure provider: %v", err)
	}

	if _, err := svc.Authenticate(AuthRequest{Method: AuthMethodOAuth, Username: "", Secret: "abcdefghi"}); err == nil {
		t.Fatalf("expected validation error for username")
	}

	if _, err := svc.Authenticate(AuthRequest{Method: AuthMethodOAuth, Username: "ops@example.com", Secret: "short"}); err == nil {
		t.Fatalf("expected validation error for secret length")
	}
}
