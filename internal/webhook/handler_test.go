package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/andrelair-platform/minicloud-plane/internal/plane"
)

// mockPublisher records calls and can be wired to return an error.
type mockPublisher struct {
	calls []publishCall
	err   error
}

type publishCall struct {
	event  string
	action string
}

func (m *mockPublisher) Publish(_ context.Context, event, action string, _ any) error {
	m.calls = append(m.calls, publishCall{event: event, action: action})
	return m.err
}

func signBody(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

func TestServeHTTP_MethodNotAllowed(t *testing.T) {
	h := NewHandler("", &mockPublisher{})
	req := httptest.NewRequest(http.MethodGet, "/webhook", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestServeHTTP_NoSecretPublishes(t *testing.T) {
	pub := &mockPublisher{}
	h := NewHandler("", pub)

	event := plane.WebhookEvent{Event: "issue", Action: "created", Actor: "kanmegnea"}
	body, _ := json.Marshal(event)
	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if len(pub.calls) != 1 {
		t.Fatalf("expected 1 publish call, got %d", len(pub.calls))
	}
	if pub.calls[0].event != "issue" || pub.calls[0].action != "created" {
		t.Errorf("unexpected publish: %+v", pub.calls[0])
	}
}

func TestServeHTTP_ValidSignature(t *testing.T) {
	secret := "test-secret"
	pub := &mockPublisher{}
	h := NewHandler(secret, pub)

	event := plane.WebhookEvent{Event: "cycle", Action: "updated"}
	body, _ := json.Marshal(event)
	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	req.Header.Set("X-Plane-Signature", signBody(secret, body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if len(pub.calls) != 1 {
		t.Error("expected publish to be called once")
	}
}

func TestServeHTTP_InvalidSignature(t *testing.T) {
	h := NewHandler("real-secret", &mockPublisher{})
	body := []byte(`{"event":"issue","action":"deleted"}`)
	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	req.Header.Set("X-Plane-Signature", "deadbeef")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestServeHTTP_InvalidJSON(t *testing.T) {
	h := NewHandler("", &mockPublisher{})
	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader([]byte("not-json")))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestServeHTTP_PublishErrorDoesNotFail(t *testing.T) {
	// NATS failure must not cause a 5xx (Plane would retry and loop).
	pub := &mockPublisher{err: errors.New("nats down")}
	h := NewHandler("", pub)

	event := plane.WebhookEvent{Event: "module", Action: "created"}
	body, _ := json.Marshal(event)
	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 even on NATS error, got %d", rec.Code)
	}
}

func TestVerifySignature_ValidAndInvalid(t *testing.T) {
	h := &Handler{secret: "key"}
	body := []byte("payload")
	if !h.verifySignature(signBody("key", body), body) {
		t.Error("valid signature should pass")
	}
	if h.verifySignature("badsig", body) {
		t.Error("bad signature should fail")
	}
}
