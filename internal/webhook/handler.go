package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"

	natspub "github.com/andrelair-platform/minicloud-plane/internal/nats"
	"github.com/andrelair-platform/minicloud-plane/internal/plane"
)

type Handler struct {
	secret    string
	publisher *natspub.Publisher
}

func NewHandler(secret string, publisher *natspub.Publisher) *Handler {
	return &Handler{secret: secret, publisher: publisher}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		http.Error(w, "read error", http.StatusBadRequest)
		return
	}

	if h.secret != "" && !h.verifySignature(r.Header.Get("X-Plane-Signature"), body) {
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	var event plane.WebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("webhook: event=%s action=%s actor=%s", event.Event, event.Action, event.Actor)

	if err := h.publisher.Publish(event.Event, event.Action, event); err != nil {
		log.Printf("WARN: NATS publish failed: %v", err)
		// Don't return 500 — Plane retries on failure and we'd loop.
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) verifySignature(sig string, body []byte) bool {
	mac := hmac.New(sha256.New, []byte(h.secret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(sig), []byte(expected))
}
