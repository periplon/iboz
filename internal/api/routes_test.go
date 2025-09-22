package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/example/iboz/internal/email"
	"github.com/example/iboz/internal/email/adapter/memory"
	"github.com/example/iboz/internal/email/adapter/synthetic"
)

type stubEmailService struct{}

func (stubEmailService) ConfigureProvider(context.Context, email.ProviderConfig) error { return nil }
func (stubEmailService) Authenticate(context.Context, email.AuthRequest) (email.AuthState, error) {
	return email.AuthState{}, nil
}
func (stubEmailService) FetchEmails(context.Context) ([]email.EmailMessage, error) { return nil, nil }
func (stubEmailService) State(context.Context) (email.ServiceState, error) {
	return email.ServiceState{}, nil
}

type testClock struct {
	now time.Time
}

func (t testClock) Now() time.Time {
	return t.now
}

func newEmailHandler(t *testing.T) handler {
	t.Helper()
	repo := memory.NewRepository()
	clock := testClock{now: time.Date(2025, time.March, 18, 12, 0, 0, 0, time.UTC)}
	svc := email.NewService(repo, email.NewSHA256Hasher(), synthetic.NewGenerator(), clock)
	return handler{emailService: svc}
}

func TestRegisterRegistersExpectedRoutes(t *testing.T) {
	e := echo.New()
	Register(e.Group("/api"), stubEmailService{})

	expected := map[string]bool{
		http.MethodGet + "/api/health":                       true,
		http.MethodGet + "/api/dashboard":                    true,
		http.MethodGet + "/api/focus/plan":                   true,
		http.MethodGet + "/api/automations":                  true,
		http.MethodPost + "/api/automations/test-run":        true,
		http.MethodGet + "/api/email/provider":               true,
		http.MethodPost + "/api/email/provider":              true,
		http.MethodPost + "/api/email/provider/authenticate": true,
		http.MethodGet + "/api/email/messages":               true,
	}

	for _, route := range e.Routes() {
		key := route.Method + route.Path
		delete(expected, key)
	}

	if len(expected) != 0 {
		t.Fatalf("missing routes: %v", expected)
	}
}

func newContext(method, target string, body *bytes.Buffer) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, target, body)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	} else {
		req = httptest.NewRequest(method, target, nil)
	}
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func decodeBody[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	var result T
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	return result
}

func TestHealthHandler(t *testing.T) {
	ctx, rec := newContext(http.MethodGet, "/api/health", nil)

	if err := healthHandler(ctx); err != nil {
		t.Fatalf("health handler returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rec.Code)
	}

	resp := decodeBody[map[string]any](t, rec)

	if resp["status"] != "ok" {
		t.Fatalf("unexpected status payload: %v", resp["status"])
	}

	if _, err := time.Parse(time.RFC3339, resp["timestamp"].(string)); err != nil {
		t.Fatalf("timestamp was not RFC3339: %v", err)
	}
}

func TestDashboardHandler(t *testing.T) {
	ctx, rec := newContext(http.MethodGet, "/api/dashboard", nil)

	if err := dashboardHandler(ctx); err != nil {
		t.Fatalf("dashboard handler returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rec.Code)
	}

	resp := decodeBody[map[string]any](t, rec)

	summary, ok := resp["summary"].(map[string]any)
	if !ok {
		t.Fatalf("summary not found in response: %v", resp)
	}
	if _, ok := summary["inboxZeroTarget"]; !ok {
		t.Fatalf("summary missing inboxZeroTarget: %v", summary)
	}

	queues, ok := resp["queues"].([]any)
	if !ok || len(queues) != 4 {
		t.Fatalf("expected four queues, got %v", resp["queues"])
	}

	recommendations, ok := resp["recommendations"].([]any)
	if !ok || len(recommendations) != 2 {
		t.Fatalf("expected two recommendations, got %v", resp["recommendations"])
	}

	focusSessions, ok := resp["focusSessions"].([]any)
	if !ok || len(focusSessions) != 2 {
		t.Fatalf("expected two focus sessions, got %v", resp["focusSessions"])
	}
}

func TestFocusPlanHandler(t *testing.T) {
	ctx, rec := newContext(http.MethodGet, "/api/focus/plan", nil)

	if err := focusPlanHandler(ctx); err != nil {
		t.Fatalf("focus plan handler returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rec.Code)
	}

	resp := decodeBody[map[string]any](t, rec)

	if _, err := time.Parse(time.RFC3339, resp["date"].(string)); err != nil {
		t.Fatalf("date was not RFC3339: %v", err)
	}

	sessions, ok := resp["sessions"].([]any)
	if !ok || len(sessions) != 2 {
		t.Fatalf("expected two sessions, got %v", resp["sessions"])
	}

	metrics, ok := resp["metrics"].(map[string]any)
	if !ok || len(metrics) == 0 {
		t.Fatalf("metrics missing or empty: %v", resp["metrics"])
	}

	controls, ok := resp["controls"].(map[string]any)
	if !ok || len(controls) == 0 {
		t.Fatalf("controls missing or empty: %v", resp["controls"])
	}
}

func TestAutomationsHandler(t *testing.T) {
	ctx, rec := newContext(http.MethodGet, "/api/automations", nil)

	if err := automationsHandler(ctx); err != nil {
		t.Fatalf("automations handler returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rec.Code)
	}

	resp := decodeBody[map[string]any](t, rec)

	templates, ok := resp["templates"].([]any)
	if !ok || len(templates) != 2 {
		t.Fatalf("expected two templates, got %v", resp["templates"])
	}

	if _, ok := resp["overview"].(map[string]any); !ok {
		t.Fatalf("overview missing: %v", resp)
	}
}

func TestAutomationTestRunHandlerSuccess(t *testing.T) {
	payload := map[string]any{
		"templateId": "auto-ack",
		"parameters": map[string]any{"foo": "bar"},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	ctx, rec := newContext(http.MethodPost, "/api/automations/test-run", bytes.NewBuffer(body))

	if err := automationTestRunHandler(ctx); err != nil {
		t.Fatalf("automation test run handler returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rec.Code)
	}

	resp := decodeBody[map[string]any](t, rec)

	if resp["templateId"] != "auto-ack" {
		t.Fatalf("templateId mismatch: %v", resp["templateId"])
	}

	review, ok := resp["review"].(map[string]any)
	if !ok {
		t.Fatalf("review missing: %v", resp)
	}

	if review["requiresApproval"] != true {
		t.Fatalf("expected requiresApproval to be true, got %v", review["requiresApproval"])
	}
}

func TestAutomationTestRunHandlerInvalidJSON(t *testing.T) {
	ctx, rec := newContext(http.MethodPost, "/api/automations/test-run", bytes.NewBufferString("{"))

	if err := automationTestRunHandler(ctx); err != nil {
		t.Fatalf("automation test run handler returned error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request status, got %d", rec.Code)
	}
}

func TestAutomationTestRunHandlerMissingTemplateID(t *testing.T) {
	payload := map[string]any{
		"parameters": map[string]any{"foo": "bar"},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	ctx, rec := newContext(http.MethodPost, "/api/automations/test-run", bytes.NewBuffer(body))

	if err := automationTestRunHandler(ctx); err != nil {
		t.Fatalf("automation test run handler returned error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request status, got %d", rec.Code)
	}
}

func TestEmailProviderHandlers(t *testing.T) {
	h := newEmailHandler(t)

	ctx, rec := newContext(http.MethodGet, "/api/email/provider", nil)
	if err := h.emailProviderStateHandler(ctx); err != nil {
		t.Fatalf("email state handler error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rec.Code)
	}

	state := decodeBody[emailProviderStateResponse](t, rec)
	if state.MessagesFetched != 0 {
		t.Fatalf("expected zero fetched messages, got %d", state.MessagesFetched)
	}

	cfg := map[string]any{
		"provider":    "gmail",
		"displayName": "Pilot Inbox",
		"connection": map[string]any{
			"protocol": "api",
		},
		"syncWindowHours": 48,
		"labelFilters":    []string{"Urgent", "Automation"},
	}
	body, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal config: %v", err)
	}

	ctx, rec = newContext(http.MethodPost, "/api/email/provider", bytes.NewBuffer(body))
	if err := h.emailProviderConfigureHandler(ctx); err != nil {
		t.Fatalf("configure handler error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rec.Code)
	}

	state = decodeBody[emailProviderStateResponse](t, rec)
	if state.Config == nil || state.Config.DisplayName != "Pilot Inbox" {
		t.Fatalf("unexpected config state: %+v", state.Config)
	}

	authPayload := map[string]any{
		"method":      "appPassword",
		"username":    "pilot@example.com",
		"appPassword": "supersafesecret",
	}
	body, err = json.Marshal(authPayload)
	if err != nil {
		t.Fatalf("marshal auth: %v", err)
	}

	ctx, rec = newContext(http.MethodPost, "/api/email/provider/authenticate", bytes.NewBuffer(body))
	if err := h.emailProviderAuthenticateHandler(ctx); err != nil {
		t.Fatalf("authenticate handler error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rec.Code)
	}

	state = decodeBody[emailProviderStateResponse](t, rec)
	if state.Auth == nil || state.Auth.Status != "connected" {
		t.Fatalf("authentication state missing or incorrect: %+v", state.Auth)
	}

	ctx, rec = newContext(http.MethodGet, "/api/email/messages", nil)
	if err := h.emailFetchMessagesHandler(ctx); err != nil {
		t.Fatalf("fetch messages handler error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rec.Code)
	}

	messages := decodeBody[emailMessagesResponse](t, rec)
	if len(messages.Messages) == 0 {
		t.Fatalf("expected synthesized messages")
	}
	if messages.SyncedAt == nil {
		t.Fatalf("expected synced timestamp")
	}
}

func TestEmailProviderAuthenticateRequiresSecret(t *testing.T) {
	h := newEmailHandler(t)

	cfgPayload := map[string]any{
		"provider":    "gmail",
		"displayName": "Ops",
		"connection":  map[string]any{"protocol": "api"},
	}
	body, err := json.Marshal(cfgPayload)
	if err != nil {
		t.Fatalf("marshal config: %v", err)
	}

	ctx, rec := newContext(http.MethodPost, "/api/email/provider", bytes.NewBuffer(body))
	if err := h.emailProviderConfigureHandler(ctx); err != nil {
		t.Fatalf("configure provider: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rec.Code)
	}

	ctx, rec = newContext(http.MethodPost, "/api/email/provider/authenticate", bytes.NewBufferString(`{"method":"oauth","username":"ops@example.com"}`))
	if err := h.emailProviderAuthenticateHandler(ctx); err != nil {
		t.Fatalf("authenticate handler error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request, got %d", rec.Code)
	}
}
