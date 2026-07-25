package plane

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProjects_OK(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Api-Key") != "test-token" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		json.NewEncoder(w).Encode(ListResponse[Project]{
			Count:   1,
			Results: []Project{{ID: "proj-1", Name: "Platform Demo", Identifier: "PT"}},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", "myworkspace")
	projects, err := client.Projects()
	if err != nil {
		t.Fatalf("Projects() error: %v", err)
	}
	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}
	if projects[0].ID != "proj-1" || projects[0].Name != "Platform Demo" {
		t.Errorf("unexpected project: %+v", projects[0])
	}
}

func TestProjects_SendsAuthHeader(t *testing.T) {
	var gotKey string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.Header.Get("X-Api-Key")
		json.NewEncoder(w).Encode(ListResponse[Project]{})
	}))
	defer server.Close()

	NewClient(server.URL, "my-secret-token", "ws").Projects() //nolint:errcheck
	if gotKey != "my-secret-token" {
		t.Errorf("expected X-Api-Key=my-secret-token, got %q", gotKey)
	}
}

func TestIssues_OK(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(ListResponse[Issue]{
			Count: 2,
			Results: []Issue{
				{ID: "i1", Name: "Fix bug", Priority: "urgent"},
				{ID: "i2", Name: "Add feature", Priority: "low"},
			},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "tok", "ws")
	issues, err := client.Issues("proj-1")
	if err != nil {
		t.Fatalf("Issues() error: %v", err)
	}
	if len(issues) != 2 {
		t.Errorf("expected 2 issues, got %d", len(issues))
	}
	if issues[0].Name != "Fix bug" {
		t.Errorf("unexpected issue name: %q", issues[0].Name)
	}
}

func TestIssue_OK(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(Issue{ID: "i42", Name: "Single issue"})
	}))
	defer server.Close()

	client := NewClient(server.URL, "tok", "ws")
	issue, err := client.Issue("proj-1", "i42")
	if err != nil {
		t.Fatalf("Issue() error: %v", err)
	}
	if issue.ID != "i42" {
		t.Errorf("unexpected issue ID: %q", issue.ID)
	}
}

func TestClient_HTTP4xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient(server.URL, "tok", "ws")
	_, err := client.Projects()
	if err == nil {
		t.Error("expected error for HTTP 404, got nil")
	}
}

func TestClient_HTTP5xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL, "tok", "ws")
	_, err := client.Issues("proj-1")
	if err == nil {
		t.Error("expected error for HTTP 500, got nil")
	}
}
