//go:build integration

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/andrelair-platform/minicloud-plane/internal/plane"
)

// ── test doubles ─────────────────────────────────────────────────────────────

type mockPlane struct {
	projects []plane.Project
	issues   map[string][]plane.Issue
}

func (m *mockPlane) Projects() ([]plane.Project, error) { return m.projects, nil }
func (m *mockPlane) Issues(id string) ([]plane.Issue, error) {
	return m.issues[id], nil
}

type mockPub struct{ mu sync.Mutex; n int }

func (m *mockPub) Publish(_ context.Context, _, _ string, _ any) error {
	m.mu.Lock()
	m.n++
	m.mu.Unlock()
	return nil
}

// ── shared server fixture ────────────────────────────────────────────────────

var (
	srv     *httptest.Server
	pub     *mockPub
	planeDB *mockPlane
	once    sync.Once
)

func setupServer(t *testing.T) string {
	t.Helper()
	once.Do(func() {
		pub = &mockPub{}
		planeDB = &mockPlane{
			projects: []plane.Project{{ID: "p1", Name: "Platform"}},
			issues:   map[string][]plane.Issue{"p1": {{ID: "i1", Name: "Bug"}}},
		}
		srv = httptest.NewServer(newServer(planeDB, pub, ""))
	})
	return srv.URL
}

func doGet(t *testing.T, path string) *http.Response {
	t.Helper()
	resp, err := http.Get(setupServer(t) + path)
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
	return resp
}

func doPost(t *testing.T, path string, body []byte, headers map[string]string) *http.Response {
	t.Helper()
	req, _ := http.NewRequest(http.MethodPost, setupServer(t)+path, bytes.NewReader(body))
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST %s: %v", path, err)
	}
	return resp
}

// ── tests ────────────────────────────────────────────────────────────────────

func TestIntegration_Health(t *testing.T) {
	resp := doGet(t, "/health")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Errorf("unexpected Content-Type: %q", ct)
	}
	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode health: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("status=%q, want ok", body["status"])
	}
	if body["service"] != "minicloud-plane" {
		t.Errorf("service=%q, want minicloud-plane", body["service"])
	}
}

func TestIntegration_APIProjects(t *testing.T) {
	resp := doGet(t, "/api/projects")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	var projects []plane.Project
	if err := json.NewDecoder(resp.Body).Decode(&projects); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(projects) != 1 || projects[0].ID != "p1" {
		t.Errorf("unexpected projects: %+v", projects)
	}
}

func TestIntegration_APIIssues(t *testing.T) {
	resp := doGet(t, "/api/projects/p1/issues")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	var issues []plane.Issue
	if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(issues) != 1 || issues[0].ID != "i1" {
		t.Errorf("unexpected issues: %+v", issues)
	}
}

func TestIntegration_WebhookPublishes(t *testing.T) {
	before := pub.n
	payload, _ := json.Marshal(map[string]string{
		"event": "issue", "action": "created", "actor": "test",
	})
	resp := doPost(t, "/webhook", payload, nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if pub.n != before+1 {
		t.Errorf("expected 1 publish call, pub.n went from %d to %d", before, pub.n)
	}
}

func TestIntegration_Metrics(t *testing.T) {
	doGet(t, "/health").Body.Close()

	resp := doGet(t, "/metrics")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "http_requests_total") {
		t.Error("metrics missing http_requests_total")
	}
}

func TestIntegration_APIUnknownPath(t *testing.T) {
	resp := doGet(t, "/api/unknown/path")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}
