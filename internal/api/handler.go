package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/andrelair-platform/minicloud-plane/internal/plane"
)

// Handler exposes Plane data over a simple REST API.
// Consumed by Backstage and other internal tools.
type Handler struct {
	client *plane.Client
}

func NewHandler(client *plane.Client) *Handler {
	return &Handler{client: client}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	path := strings.TrimPrefix(r.URL.Path, "/api")
	switch {
	case path == "/projects" || path == "/projects/":
		h.projects(w, r)
	case strings.HasPrefix(path, "/projects/") && strings.HasSuffix(path, "/issues"):
		parts := strings.Split(strings.Trim(path, "/"), "/")
		if len(parts) == 3 {
			h.issues(w, r, parts[1])
		}
	default:
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
	}
}

func (h *Handler) projects(w http.ResponseWriter, _ *http.Request) {
	projects, err := h.client.Projects()
	if err != nil {
		log.Printf("projects error: %v", err)
		http.Error(w, `{"error":"upstream error"}`, http.StatusBadGateway)
		return
	}
	json.NewEncoder(w).Encode(projects)
}

func (h *Handler) issues(w http.ResponseWriter, _ *http.Request, projectID string) {
	issues, err := h.client.Issues(projectID)
	if err != nil {
		log.Printf("issues error: %v", err)
		http.Error(w, `{"error":"upstream error"}`, http.StatusBadGateway)
		return
	}
	json.NewEncoder(w).Encode(issues)
}
