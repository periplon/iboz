package synthetic

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/example/iboz/internal/email"
)

var _ email.MessageGenerator = (*Generator)(nil)

// Generator produces deterministic sample messages for use in tests and demos.
type Generator struct{}

// NewGenerator constructs a new synthetic message generator.
func NewGenerator() *Generator {
	return &Generator{}
}

// Generate returns a deterministic set of email messages for the supplied configuration.
func (g *Generator) Generate(ctx context.Context, cfg email.ProviderConfig, auth email.AuthState, now time.Time) ([]email.EmailMessage, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	baseLabels := append([]string{"INBOX"}, cfg.LabelFilters...)

	summary := email.EmailMessage{
		ID:         "msg-schedule",
		Subject:    fmt.Sprintf("%s focus queue summary", strings.Title(cfg.DisplayName)),
		Sender:     fmt.Sprintf("notifications@%s", cfg.Provider),
		ReceivedAt: now.Add(-45 * time.Minute),
		Snippet:    fmt.Sprintf("Automation insights for %s. 3 urgent items need review.", cfg.DisplayName),
		Labels:     append([]string(nil), baseLabels...),
		Importance: "high",
	}

	escalated := email.EmailMessage{
		ID:         "msg-escalation",
		Subject:    "Escalation: Contract signature pending",
		Sender:     "legal-ops@example.com",
		ReceivedAt: now.Add(-2 * time.Hour),
		Snippet:    fmt.Sprintf("Hi %s, procurement is awaiting countersignature from vendor.", auth.Username),
		Labels:     append([]string{"Escalations"}, cfg.LabelFilters...),
		Importance: "high",
	}

	digest := email.EmailMessage{
		ID:         "msg-digest",
		Subject:    "Daily automation digest",
		Sender:     "automation-bot@example.com",
		ReceivedAt: now.Add(-6 * time.Hour),
		Snippet:    fmt.Sprintf("%d workflows executed, 12 emails triaged automatically.", 4+cfg.SyncWindowHours/24),
		Labels:     append([]string{"Automation"}, cfg.LabelFilters...),
		Importance: "normal",
	}

	return []email.EmailMessage{summary, escalated, digest}, nil
}
