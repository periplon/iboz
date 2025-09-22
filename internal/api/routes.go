package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/example/iboz/internal/email"
)

type handler struct {
	emailService email.ProviderService
}

// Register wires the API routes to the provided echo group.
func Register(g *echo.Group, emailService email.ProviderService) {
	if emailService == nil {
		panic("api: email service dependency is required")
	}
	h := handler{emailService: emailService}

	g.GET("/health", healthHandler)
	g.GET("/dashboard", dashboardHandler)
	g.GET("/focus/plan", focusPlanHandler)
	g.GET("/automations", automationsHandler)
	g.POST("/automations/test-run", automationTestRunHandler)

	emailGroup := g.Group("/email")
	emailGroup.GET("/provider", h.emailProviderStateHandler)
	emailGroup.POST("/provider", h.emailProviderConfigureHandler)
	emailGroup.POST("/provider/authenticate", h.emailProviderAuthenticateHandler)
	emailGroup.GET("/messages", h.emailFetchMessagesHandler)
}

func healthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().UTC(),
	})
}

func dashboardHandler(c echo.Context) error {
	payload := map[string]interface{}{
		"summary": map[string]interface{}{
			"inboxZeroTarget":  10,
			"currentInbox":     7,
			"automationRate":   0.68,
			"timeSavedMinutes": 125,
		},
		"focusSessions": []map[string]interface{}{
			{
				"id":          "focus-urgent",
				"label":       "Urgent action block",
				"start":       "09:30",
				"estimated":   45,
				"emails":      4,
				"llmSupport":  true,
				"description": "Handle escalations with recommended replies and audit log links.",
			},
			{
				"id":          "focus-followup",
				"label":       "Delegated follow-ups",
				"start":       "14:00",
				"estimated":   30,
				"emails":      6,
				"llmSupport":  false,
				"description": "Batch nudges on waiting approvals before end of day.",
			},
		},
		"queues": []map[string]interface{}{
			{
				"id":          "urgent",
				"label":       "Urgent",
				"description": "Requires action within 4 hours",
				"count":       3,
				"llmEnabled":  true,
			},
			{
				"id":          "today",
				"label":       "Today",
				"description": "Recommended to clear before end of day",
				"count":       9,
				"llmEnabled":  true,
			},
			{
				"id":          "waiting",
				"label":       "Waiting",
				"description": "Awaiting responses from others",
				"count":       4,
				"llmEnabled":  false,
			},
			{
				"id":          "delegated",
				"label":       "Delegated",
				"description": "Assigned to teammates with SLA tracking",
				"count":       5,
				"llmEnabled":  false,
			},
		},
		"recommendations": []map[string]interface{}{
			{
				"id":          "digest",
				"title":       "Send weekly digest",
				"description": "Snooze newsletters into a Friday recap.",
				"confidence":  0.87,
			},
			{
				"id":          "delegate-contract",
				"title":       "Delegate contract review",
				"description": "Assign to legal ops and request update in 2 days.",
				"confidence":  0.73,
			},
		},
	}

	return c.JSON(http.StatusOK, payload)
}

func focusPlanHandler(c echo.Context) error {
	payload := map[string]interface{}{
		"date": time.Now().UTC().Format(time.RFC3339),
		"sessions": []map[string]interface{}{
			{
				"id":          "focus-urgent",
				"label":       "Urgent Action",
				"estimated":   45,
				"emails":      4,
				"llmSupport":  true,
				"description": "Batch handle urgent client escalations with recommended replies.",
			},
			{
				"id":          "focus-followup",
				"label":       "Follow-Up",
				"estimated":   30,
				"emails":      6,
				"llmSupport":  false,
				"description": "Send nudges on pending approvals and delegated tasks.",
			},
		},
		"metrics": map[string]interface{}{
			"clearedToday": 18,
			"streak":       4,
			"goal":         3,
		},
		"controls": map[string]interface{}{
			"notificationsMuted": true,
			"batchingEnabled":    true,
			"autoSummaries":      true,
		},
	}

	return c.JSON(http.StatusOK, payload)
}

func automationsHandler(c echo.Context) error {
	payload := map[string]interface{}{
		"overview": map[string]interface{}{
			"active":             12,
			"automationCoverage": 0.74,
			"avgTimeSaved":       32,
		},
		"templates": []map[string]interface{}{
			{
				"id":          "auto-ack",
				"name":        "Auto-acknowledge support tickets",
				"description": "Send branded receipt, create Jira issue, and assign to support queue.",
				"trigger":     "sender: support@customer.com",
				"conditions": []string{
					"subject CONTAINS 'case #'",
					"attachment.type = 'pdf'",
				},
				"actions": []string{
					"Send template response",
					"Create Jira ticket",
					"Label as Waiting",
				},
				"requiresApproval": true,
				"owner":            "Support Operations",
				"lastRun":          "2025-03-18T10:24:00Z",
			},
			{
				"id":          "vip-sms",
				"name":        "VIP Escalation to Slack",
				"description": "If VIP contacts after hours, alert on-call channel and schedule morning follow-up.",
				"trigger":     "tag:vip AND time:after_hours",
				"conditions": []string{
					"sender IN vip_list",
					"llm.confidence >= 0.65",
				},
				"actions": []string{
					"Post to #escalations channel",
					"Create Asana task",
					"Snooze email until 8am",
				},
				"requiresApproval": false,
				"owner":            "Customer Experience",
				"lastRun":          "2025-03-18T07:10:00Z",
			},
		},
	}

	return c.JSON(http.StatusOK, payload)
}

type emailProviderStateResponse struct {
	Config          *email.ProviderConfig `json:"config,omitempty"`
	Auth            *email.AuthState      `json:"auth,omitempty"`
	LastSync        *string               `json:"lastSync,omitempty"`
	MessagesFetched int                   `json:"messagesFetched"`
}

type emailAuthRequest struct {
	Method      email.AuthMethod `json:"method"`
	Username    string           `json:"username"`
	AppPassword string           `json:"appPassword"`
	OAuthToken  string           `json:"oauthToken"`
}

type emailMessagesResponse struct {
	Messages []email.EmailMessage `json:"messages"`
	SyncedAt *string              `json:"syncedAt,omitempty"`
}

func (h handler) emailProviderStateHandler(c echo.Context) error {
	return h.respondWithEmailState(c, http.StatusOK)
}

func (h handler) emailProviderConfigureHandler(c echo.Context) error {
	var cfg email.ProviderConfig
	if err := c.Bind(&cfg); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid configuration payload"})
	}

	if err := h.emailService.ConfigureProvider(c.Request().Context(), cfg); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	return h.respondWithEmailState(c, http.StatusOK)
}

func (h handler) emailProviderAuthenticateHandler(c echo.Context) error {
	var req emailAuthRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid authentication payload"})
	}

	var secret string
	switch req.Method {
	case email.AuthMethodAppPassword:
		secret = req.AppPassword
	case email.AuthMethodOAuth:
		secret = req.OAuthToken
	default:
		return c.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("unsupported auth method %q", req.Method)})
	}

	if secret == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "credential secret is required"})
	}

	if _, err := h.emailService.Authenticate(c.Request().Context(), email.AuthRequest{Method: req.Method, Username: req.Username, Secret: secret}); err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, email.ErrProviderNotConfigured):
			status = http.StatusBadRequest
		default:
			status = http.StatusInternalServerError
		}
		return c.JSON(status, map[string]string{"error": err.Error()})
	}

	return h.respondWithEmailState(c, http.StatusOK)
}

func (h handler) emailFetchMessagesHandler(c echo.Context) error {
	messages, err := h.emailService.FetchEmails(c.Request().Context())
	if err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, email.ErrProviderNotConfigured), errors.Is(err, email.ErrProviderNotAuthenticated):
			status = http.StatusBadRequest
		case errors.Is(err, context.Canceled):
			status = http.StatusRequestTimeout
		}
		return c.JSON(status, map[string]string{"error": err.Error()})
	}

	state, err := h.emailService.State(c.Request().Context())
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, context.Canceled) {
			status = http.StatusRequestTimeout
		}
		return c.JSON(status, map[string]string{"error": err.Error()})
	}

	var syncedAt *string
	if !state.LastSync.IsZero() {
		formatted := state.LastSync.Format(time.RFC3339)
		syncedAt = &formatted
	}

	return c.JSON(http.StatusOK, emailMessagesResponse{
		Messages: messages,
		SyncedAt: syncedAt,
	})
}

func (h handler) respondWithEmailState(c echo.Context, status int) error {
	state, err := h.emailService.State(c.Request().Context())
	if err != nil {
		httpStatus := http.StatusInternalServerError
		if errors.Is(err, context.Canceled) {
			httpStatus = http.StatusRequestTimeout
		}
		return c.JSON(httpStatus, map[string]string{"error": err.Error()})
	}

	var lastSync *string
	if !state.LastSync.IsZero() {
		formatted := state.LastSync.Format(time.RFC3339)
		lastSync = &formatted
	}

	return c.JSON(status, emailProviderStateResponse{
		Config:          state.Config,
		Auth:            state.Auth,
		LastSync:        lastSync,
		MessagesFetched: state.MessagesFetched,
	})
}

func automationTestRunHandler(c echo.Context) error {
	var input struct {
		TemplateID string                 `json:"templateId"`
		Parameters map[string]interface{} `json:"parameters"`
	}

	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid payload"})
	}

	if input.TemplateID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "templateId is required"})
	}

	response := map[string]interface{}{
		"templateId": input.TemplateID,
		"status":     "simulated",
		"summary":    "Automation would execute 3 actions with estimated savings of 12 minutes.",
		"parameters": input.Parameters,
		"review": map[string]interface{}{
			"requiresApproval": input.TemplateID == "auto-ack",
			"confidence":       0.82,
		},
	}

	return c.JSON(http.StatusOK, response)
}
