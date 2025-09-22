package memory

import (
	"context"
	"sync"
	"time"

	"github.com/example/iboz/internal/email"
)

var _ email.Repository = (*Repository)(nil)

// Repository provides an in-memory implementation of the email.Repository port.
type Repository struct {
	mu       sync.RWMutex
	config   *email.ProviderConfig
	auth     *email.AuthRecord
	messages []email.EmailMessage
	lastSync time.Time
}

// NewRepository builds a new in-memory repository instance.
func NewRepository() *Repository {
	return &Repository{}
}

// SaveConfig stores the provider configuration.
func (r *Repository) SaveConfig(ctx context.Context, cfg email.ProviderConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	clone := cfg
	clone.LabelFilters = append([]string(nil), cfg.LabelFilters...)

	r.mu.Lock()
	r.config = &clone
	r.mu.Unlock()

	return nil
}

// GetConfig returns the stored provider configuration if present.
func (r *Repository) GetConfig(ctx context.Context) (*email.ProviderConfig, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.config == nil {
		return nil, nil
	}

	clone := *r.config
	clone.LabelFilters = append([]string(nil), r.config.LabelFilters...)
	return &clone, nil
}

// ClearAuth removes any stored authentication metadata.
func (r *Repository) ClearAuth(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	r.auth = nil
	r.mu.Unlock()
	return nil
}

// SaveAuth persists the authentication record.
func (r *Repository) SaveAuth(ctx context.Context, record email.AuthRecord) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	clone := record
	clone.State.UpdatedAt = record.State.UpdatedAt

	r.mu.Lock()
	r.auth = &clone
	r.mu.Unlock()
	return nil
}

// GetAuth retrieves the stored authentication record if present.
func (r *Repository) GetAuth(ctx context.Context) (*email.AuthRecord, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.auth == nil {
		return nil, nil
	}

	clone := *r.auth
	clone.State = r.auth.State
	return &clone, nil
}

// ClearMessages removes cached messages and last sync metadata.
func (r *Repository) ClearMessages(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	r.messages = nil
	r.lastSync = time.Time{}
	r.mu.Unlock()
	return nil
}

// SaveMessages replaces the cached messages and updates the last sync timestamp.
func (r *Repository) SaveMessages(ctx context.Context, messages []email.EmailMessage, syncedAt time.Time) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	cloned := cloneMessages(messages)

	r.mu.Lock()
	r.messages = cloned
	r.lastSync = syncedAt
	r.mu.Unlock()
	return nil
}

// GetMessages returns the cached messages and associated last sync timestamp.
func (r *Repository) GetMessages(ctx context.Context) ([]email.EmailMessage, time.Time, error) {
	if err := ctx.Err(); err != nil {
		return nil, time.Time{}, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.messages) == 0 {
		return nil, r.lastSync, nil
	}

	return cloneMessages(r.messages), r.lastSync, nil
}

func cloneMessages(messages []email.EmailMessage) []email.EmailMessage {
	if len(messages) == 0 {
		return nil
	}
	cloned := make([]email.EmailMessage, len(messages))
	for i, msg := range messages {
		cloned[i] = msg
		cloned[i].Labels = append([]string(nil), msg.Labels...)
	}
	return cloned
}
