package email

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

// Supported provider identifiers.
const (
	ProviderGmail   = "gmail"
	ProviderOutlook = "outlook"
	ProviderIMAP    = "imap"

	ProtocolAPI  = "api"
	ProtocolIMAP = "imap"
)

var (
	supportedProviders = map[string]struct{}{
		ProviderGmail:   {},
		ProviderOutlook: {},
		ProviderIMAP:    {},
	}
	supportedProtocols = map[string]struct{}{
		ProtocolAPI:  {},
		ProtocolIMAP: {},
	}
)

var (
	// ErrProviderNotConfigured is returned when configuration is missing.
	ErrProviderNotConfigured = errors.New("email provider not configured")
	// ErrProviderNotAuthenticated is returned when authentication is missing.
	ErrProviderNotAuthenticated = errors.New("email provider authentication not configured")
)

// ConnectionSettings describes how to reach the upstream provider.
type ConnectionSettings struct {
	Protocol string `json:"protocol"`
	Host     string `json:"host,omitempty"`
	Port     int    `json:"port,omitempty"`
	UseTLS   bool   `json:"useTls,omitempty"`
	APIBase  string `json:"apiBaseUrl,omitempty"`
}

// ProviderConfig captures provider level configuration.
type ProviderConfig struct {
	Provider        string             `json:"provider"`
	DisplayName     string             `json:"displayName"`
	Connection      ConnectionSettings `json:"connection"`
	SyncWindowHours int                `json:"syncWindowHours"`
	LabelFilters    []string           `json:"labelFilters"`
}

// AuthMethod enumerates supported credential flows.
type AuthMethod string

const (
	AuthMethodOAuth       AuthMethod = "oauth"
	AuthMethodAppPassword AuthMethod = "appPassword"
)

// AuthState reports the persisted authentication metadata.
type AuthState struct {
	Method    AuthMethod `json:"method"`
	Username  string     `json:"username,omitempty"`
	Status    string     `json:"status"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

// EmailMessage represents a fetched message from the provider.
type EmailMessage struct {
	ID         string    `json:"id"`
	Subject    string    `json:"subject"`
	Sender     string    `json:"sender"`
	ReceivedAt time.Time `json:"receivedAt"`
	Snippet    string    `json:"snippet"`
	Labels     []string  `json:"labels"`
	Importance string    `json:"importance"`
}

// ServiceState captures the public state exported by the service.
type ServiceState struct {
	Config          *ProviderConfig
	Auth            *AuthState
	LastSync        time.Time
	MessagesFetched int
}

// AuthRequest collects credentials for authentication flows.
type AuthRequest struct {
	Method   AuthMethod
	Username string
	Secret   string
}

// Service manages provider configuration, authentication and message retrieval.
type Service struct {
	mu       sync.RWMutex
	config   *ProviderConfig
	auth     *authRecord
	cached   []EmailMessage
	lastSync time.Time
}

type authRecord struct {
	state      AuthState
	secretHash string
}

// NewService constructs a fresh Service instance.
func NewService() *Service {
	return &Service{}
}

// ConfigureProvider validates and stores provider configuration.
func (s *Service) ConfigureProvider(cfg ProviderConfig) error {
	if err := validateConfig(cfg); err != nil {
		return err
	}

	cleaned := normalizeConfig(cfg)

	s.mu.Lock()
	defer s.mu.Unlock()

	s.config = &cleaned
	s.auth = nil
	s.cached = nil
	s.lastSync = time.Time{}

	return nil
}

// Authenticate validates credentials and stores authentication metadata.
func (s *Service) Authenticate(req AuthRequest) (AuthState, error) {
	s.mu.RLock()
	configured := s.config != nil
	s.mu.RUnlock()
	if !configured {
		return AuthState{}, ErrProviderNotConfigured
	}

	if req.Method != AuthMethodOAuth && req.Method != AuthMethodAppPassword {
		return AuthState{}, fmt.Errorf("unsupported auth method %q", req.Method)
	}
	if strings.TrimSpace(req.Username) == "" {
		return AuthState{}, errors.New("username is required")
	}
	if strings.TrimSpace(req.Secret) == "" {
		return AuthState{}, errors.New("credential secret is required")
	}
	if len(req.Secret) < 8 {
		return AuthState{}, errors.New("credential secret must be at least 8 characters")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.config == nil {
		return AuthState{}, ErrProviderNotConfigured
	}

	now := time.Now().UTC()
	state := AuthState{
		Method:    req.Method,
		Username:  req.Username,
		Status:    "connected",
		UpdatedAt: now,
	}

	hash := sha256.Sum256([]byte(req.Secret))
	record := &authRecord{
		state:      state,
		secretHash: hex.EncodeToString(hash[:]),
	}

	s.auth = record
	return state, nil
}

// FetchEmails simulates retrieving a slice of messages from the configured provider.
func (s *Service) FetchEmails(ctx context.Context) ([]EmailMessage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.config == nil {
		return nil, ErrProviderNotConfigured
	}
	if s.auth == nil {
		return nil, ErrProviderNotAuthenticated
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	messages := synthesizeMessages(*s.config, s.auth.state, now)
	s.cached = cloneMessages(messages)
	s.lastSync = now

	return cloneMessages(messages), nil
}

// State returns a snapshot of the current service state suitable for JSON encoding.
func (s *Service) State() ServiceState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var cfg *ProviderConfig
	if s.config != nil {
		cloned := *s.config
		cloned.LabelFilters = append([]string(nil), s.config.LabelFilters...)
		cfg = &cloned
	}

	var auth *AuthState
	if s.auth != nil {
		state := s.auth.state
		auth = &state
	}

	return ServiceState{
		Config:          cfg,
		Auth:            auth,
		LastSync:        s.lastSync,
		MessagesFetched: len(s.cached),
	}
}

func validateConfig(cfg ProviderConfig) error {
	provider := strings.ToLower(strings.TrimSpace(cfg.Provider))
	if provider == "" {
		return errors.New("provider is required")
	}
	if _, ok := supportedProviders[provider]; !ok {
		return fmt.Errorf("unsupported provider %q", cfg.Provider)
	}
	if strings.TrimSpace(cfg.DisplayName) == "" {
		return errors.New("display name is required")
	}
	protocol := strings.ToLower(strings.TrimSpace(cfg.Connection.Protocol))
	if _, ok := supportedProtocols[protocol]; !ok {
		return fmt.Errorf("unsupported connection protocol %q", cfg.Connection.Protocol)
	}
	if protocol == ProtocolIMAP {
		if strings.TrimSpace(cfg.Connection.Host) == "" {
			return errors.New("imap host is required")
		}
		if cfg.Connection.Port <= 0 {
			return errors.New("imap port must be greater than zero")
		}
	}
	if cfg.SyncWindowHours < 0 {
		return errors.New("sync window cannot be negative")
	}
	return nil
}

func normalizeConfig(cfg ProviderConfig) ProviderConfig {
	result := ProviderConfig{
		Provider:    strings.ToLower(strings.TrimSpace(cfg.Provider)),
		DisplayName: strings.TrimSpace(cfg.DisplayName),
		Connection: ConnectionSettings{
			Protocol: strings.ToLower(strings.TrimSpace(cfg.Connection.Protocol)),
			Host:     strings.TrimSpace(cfg.Connection.Host),
			Port:     cfg.Connection.Port,
			UseTLS:   cfg.Connection.UseTLS,
			APIBase:  strings.TrimSpace(cfg.Connection.APIBase),
		},
		SyncWindowHours: cfg.SyncWindowHours,
	}

	if result.SyncWindowHours == 0 {
		result.SyncWindowHours = 24
	}

	if len(cfg.LabelFilters) > 0 {
		result.LabelFilters = make([]string, 0, len(cfg.LabelFilters))
		for _, label := range cfg.LabelFilters {
			trimmed := strings.TrimSpace(label)
			if trimmed != "" {
				result.LabelFilters = append(result.LabelFilters, trimmed)
			}
		}
	}

	if result.Connection.Protocol == ProtocolIMAP && result.Connection.Port == 0 {
		result.Connection.Port = 993
	}

	return result
}

func synthesizeMessages(cfg ProviderConfig, auth AuthState, now time.Time) []EmailMessage {
	baseLabels := append([]string{"INBOX"}, cfg.LabelFilters...)

	summary := EmailMessage{
		ID:         "msg-schedule",
		Subject:    fmt.Sprintf("%s focus queue summary", strings.Title(cfg.DisplayName)),
		Sender:     fmt.Sprintf("notifications@%s", cfg.Provider),
		ReceivedAt: now.Add(-45 * time.Minute),
		Snippet:    fmt.Sprintf("Automation insights for %s. 3 urgent items need review.", cfg.DisplayName),
		Labels:     append([]string(nil), baseLabels...),
		Importance: "high",
	}

	escalated := EmailMessage{
		ID:         "msg-escalation",
		Subject:    "Escalation: Contract signature pending",
		Sender:     "legal-ops@example.com",
		ReceivedAt: now.Add(-2 * time.Hour),
		Snippet:    fmt.Sprintf("Hi %s, procurement is awaiting countersignature from vendor.", auth.Username),
		Labels:     append([]string{"Escalations"}, cfg.LabelFilters...),
		Importance: "high",
	}

	digest := EmailMessage{
		ID:         "msg-digest",
		Subject:    "Daily automation digest",
		Sender:     "automation-bot@example.com",
		ReceivedAt: now.Add(-6 * time.Hour),
		Snippet:    fmt.Sprintf("%d workflows executed, 12 emails triaged automatically.", 4+cfg.SyncWindowHours/24),
		Labels:     append([]string{"Automation"}, cfg.LabelFilters...),
		Importance: "normal",
	}

	return []EmailMessage{summary, escalated, digest}
}

func cloneMessages(messages []EmailMessage) []EmailMessage {
	if len(messages) == 0 {
		return nil
	}
	cloned := make([]EmailMessage, len(messages))
	for i, message := range messages {
		cloned[i] = message
		cloned[i].Labels = append([]string(nil), message.Labels...)
	}
	return cloned
}
