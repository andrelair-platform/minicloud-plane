package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestInstrument_PassesThroughStatusCode(t *testing.T) {
	cases := []struct {
		name string
		code int
	}{
		{"200 OK", http.StatusOK},
		{"201 Created", http.StatusCreated},
		{"400 Bad Request", http.StatusBadRequest},
		{"404 Not Found", http.StatusNotFound},
		{"500 Internal Error", http.StatusInternalServerError},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tc.code)
			})
			h := Instrument("/test", inner)
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if rec.Code != tc.code {
				t.Errorf("expected %d, got %d", tc.code, rec.Code)
			}
		})
	}
}

func TestInstrument_DefaultStatus200(t *testing.T) {
	// When inner handler writes body without WriteHeader, status should be 200.
	inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("ok")) //nolint:errcheck
	})
	h := Instrument("/health", inner)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}
