package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"kisakay/server/internal/config"
	"kisakay/server/internal/lastfm"
	"kisakay/server/internal/views"
)

func TestHandlerRootListsAPIRoutes(t *testing.T) {
	t.Parallel()

	store, err := views.NewStore(filepath.Join(t.TempDir(), "views.json"))
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
	}()

	server := NewServer(config.Config{
		ViewHashSecret:   "test-secret",
		RateLimitEnabled: false,
	}, lastfm.NewClient("", "Kisakay", &http.Client{}), store)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)

	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	var payload rootResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if payload.Name != "Kisakay API" {
		t.Fatalf("expected API name to be set, got %q", payload.Name)
	}

	if len(payload.Routes) != 2 {
		t.Fatalf("expected 2 routes, got %d", len(payload.Routes))
	}

	if payload.Routes[0].Path != "/api/lastfm" {
		t.Fatalf("expected first route to be /api/lastfm, got %q", payload.Routes[0].Path)
	}

	if payload.Routes[1].Path != "/api/views" {
		t.Fatalf("expected second route to be /api/views, got %q", payload.Routes[1].Path)
	}
}
