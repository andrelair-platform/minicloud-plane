package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/andrelair-platform/minicloud-plane/internal/plane"
)

type mockClient struct {
	projects []plane.Project
	issues   map[string][]plane.Issue
	err      error
}

func (m *mockClient) Projects() ([]plane.Project, error) {
	return m.projects, m.err
}

func (m *mockClient) Issues(projectID string) ([]plane.Issue, error) {
	return m.issues[projectID], m.err
}

func TestProjects_OK(t *testing.T) {
	client := &mockClient{
		projects: []plane.Project{{ID: "p1", Name: "Platform"}},
	}
	h := NewHandler(client)
	req := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	var got []plane.Project
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(got) != 1 || got[0].ID != "p1" {
		t.Errorf("unexpected projects: %+v", got)
	}
}

func TestProjectsTrailingSlash_OK(t *testing.T) {
	client := &mockClient{projects: []plane.Project{{ID: "p2"}}}
	h := NewHandler(client)
	req := httptest.NewRequest(http.MethodGet, "/api/projects/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestIssues_OK(t *testing.T) {
	client := &mockClient{
		issues: map[string][]plane.Issue{
			"proj-abc": {{ID: "i1", Name: "Bug"}, {ID: "i2", Name: "Feature"}},
		},
	}
	h := NewHandler(client)
	req := httptest.NewRequest(http.MethodGet, "/api/projects/proj-abc/issues", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	var got []plane.Issue
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 issues, got %d", len(got))
	}
}

func TestUnknownPath_NotFound(t *testing.T) {
	h := NewHandler(&mockClient{})
	req := httptest.NewRequest(http.MethodGet, "/api/unknown", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestProjects_UpstreamError(t *testing.T) {
	client := &mockClient{err: errors.New("plane is down")}
	h := NewHandler(client)
	req := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Errorf("expected 502, got %d", rec.Code)
	}
}

func TestIssues_UpstreamError(t *testing.T) {
	client := &mockClient{err: errors.New("plane is down")}
	h := NewHandler(client)
	req := httptest.NewRequest(http.MethodGet, "/api/projects/x/issues", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Errorf("expected 502, got %d", rec.Code)
	}
}
