package email

import (
	"context"
	"errors"
	"fmt"
	"strings"
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

// AuthRecord stores the authentication metadata alongside the hashed secret.
type AuthRecord struct {
	State      AuthState
	SecretHash string
}

// Repository defines the persistence contract required by the service.
type Repository interface {
	SaveConfig(ctx context.Context, cfg ProviderConfig) error
	GetConfig(ctx context.Context) (*ProviderConfig, error)
	ClearAuth(ctx context.Context) error
	SaveAuth(ctx context.Context, record AuthRecord) error
	GetAuth(ctx context.Context) (*AuthRecord, error)
	ClearMessages(ctx context.Context) error
	SaveMessages(ctx context.Context, messages []EmailMessage, syncedAt time.Time) error
	GetMessages(ctx context.Context) ([]EmailMessage, time.Time, error)
}

// SecretHasher abstracts hashing of sensitive credentials.
type SecretHasher interface {
	Hash(secret string) (string, error)
}

// MessageGenerator synthesizes or retrieves messages from a provider.
type MessageGenerator interface {
	Generate(ctx context.Context, cfg ProviderConfig, auth AuthState, now time.Time) ([]EmailMessage, error)
}

// Clock abstracts time retrieval to make the service testable.
type Clock interface {
	Now() time.Time
}

// ProviderService exposes the application behaviour for configuring and fetching provider data.
type ProviderService interface {
	ConfigureProvider(ctx context.Context, cfg ProviderConfig) error
	Authenticate(ctx context.Context, req AuthRequest) (AuthState, error)
	FetchEmails(ctx context.Context) ([]EmailMessage, error)
	State(ctx context.Context) (ServiceState, error)
}

var _ ProviderService = (*Service)(nil)

// Service manages provider configuration, authentication and message retrieval.
type Service struct {
	repo      Repository
	hasher    SecretHasher
	generator MessageGenerator
	clock     Clock
}

// NewService constructs a Service instance with the supplied dependencies.
func NewService(repo Repository, hasher SecretHasher, generator MessageGenerator, clock Clock) *Service {
	if repo == nil {
		panic("email: repository dependency is required")
	}
	if hasher == nil {
		panic("email: secret hasher dependency is required")
	}
	if generator == nil {
		panic("email: message generator dependency is required")
	}
	if clock == nil {
		panic("email: clock dependency is required")
	}
	return &Service{repo: repo, hasher: hasher, generator: generator, clock: clock}
}

// ConfigureProvider validates and stores provider configuration.
func (s *Service) ConfigureProvider(ctx context.Context, cfg ProviderConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := validateConfig(cfg); err != nil {
		return err
	}

	cleaned := normalizeConfig(cfg)

	if err := s.repo.SaveConfig(ctx, cleaned); err != nil {
		return err
	}
	if err := s.repo.ClearAuth(ctx); err != nil {
		return err
	}
	if err := s.repo.ClearMessages(ctx); err != nil {
		return err
	}
	return nil
}

// Authenticate validates credentials and stores authentication metadata.
func (s *Service) Authenticate(ctx context.Context, req AuthRequest) (AuthState, error) {
	if err := ctx.Err(); err != nil {
		return AuthState{}, err
	}

	cfg, err := s.repo.GetConfig(ctx)
	if err != nil {
		return AuthState{}, err
	}
	if cfg == nil {
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

	if err := ctx.Err(); err != nil {
		return AuthState{}, err
	}

	now := s.clock.Now().UTC()
	state := AuthState{
		Method:    req.Method,
		Username:  req.Username,
		Status:    "connected",
		UpdatedAt: now,
	}

	hash, err := s.hasher.Hash(req.Secret)
	if err != nil {
		return AuthState{}, err
	}
	record := AuthRecord{State: state, SecretHash: hash}

	if err := s.repo.SaveAuth(ctx, record); err != nil {
		return AuthState{}, err
	}

	return state, nil
}

// FetchEmails retrieves a slice of messages from the configured provider.
func (s *Service) FetchEmails(ctx context.Context) ([]EmailMessage, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	cfg, err := s.repo.GetConfig(ctx)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		return nil, ErrProviderNotConfigured
	}

	auth, err := s.repo.GetAuth(ctx)
	if err != nil {
		return nil, err
	}
	if auth == nil {
		return nil, ErrProviderNotAuthenticated
	}

	now := s.clock.Now().UTC()
	messages, err := s.generator.Generate(ctx, *cfg, auth.State, now)
	if err != nil {
		return nil, err
	}

	if err := s.repo.SaveMessages(ctx, messages, now); err != nil {
		return nil, err
	}

	return cloneMessages(messages), nil
}

// State returns a snapshot of the current service state suitable for JSON encoding.
func (s *Service) State(ctx context.Context) (ServiceState, error) {
	if err := ctx.Err(); err != nil {
		return ServiceState{}, err
	}

	cfg, err := s.repo.GetConfig(ctx)
	if err != nil {
		return ServiceState{}, err
	}
	auth, err := s.repo.GetAuth(ctx)
	if err != nil {
		return ServiceState{}, err
	}
	cached, lastSync, err := s.repo.GetMessages(ctx)
	if err != nil {
		return ServiceState{}, err
	}

	return ServiceState{
		Config:          cloneProviderConfig(cfg),
		Auth:            cloneAuthState(auth),
		LastSync:        lastSync,
		MessagesFetched: len(cached),
	}, nil
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

func cloneProviderConfig(cfg *ProviderConfig) *ProviderConfig {
	if cfg == nil {
		return nil
	}
	cloned := *cfg
	cloned.LabelFilters = append([]string(nil), cfg.LabelFilters...)
	return &cloned
}

func cloneAuthState(record *AuthRecord) *AuthState {
	if record == nil {
		return nil
	}
	state := record.State
	return &state
}
