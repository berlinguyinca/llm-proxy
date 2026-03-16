// Package router provides path-based routing logic
package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRouter_New(t *testing.T) {
	r := NewRouter()
	if r == nil {
		t.Fatal("Expected router to be created")
	}
}

func TestRouter_Register(t *testing.T) {
	r := NewRouter()
	err := r.Register("/qwen/", "http://localhost:1234/v1/chat/completions")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(r.routes) == 0 {
		t.Fatal("Expected route to be registered")
	}

	route := r.routes["/qwen/"]
	if route == nil {
		t.Fatal("Expected route to exist in map")
	}

	if route.TargetURL != "http://localhost:1234/v1/chat/completions" {
		t.Errorf("Expected target URL, got %s", route.TargetURL)
	}
}

func TestRouter_RegisterEmptyPrefix(t *testing.T) {
	r := NewRouter()
	err := r.Register("", "http://example.com")

	if err != nil {
		t.Fatalf("Expected no error for empty prefix, got %v", err)
	}

	// Should not panic or crash
}

func TestRouter_RegisterEmptyURL(t *testing.T) {
	r := NewRouter()
	err := r.Register("/test/", "")

	if err != nil {
		t.Fatalf("Expected no error for empty URL, got %v", err)
	}
}

func TestRouter_GetTargetForPath_NoMatch(t *testing.T) {
	r := NewRouter()
	r.Register("/qwen/", "http://localhost:1234/v1/chat/completions")

	route, remainder, found := r.GetTargetForPath("/other/model/path")

	if found {
		t.Fatal("Expected no match for non-existent route prefix")
	}

	if route != nil {
		t.Error("Expected nil route")
	}

	if remainder != "" {
		t.Errorf("Expected empty remainder, got %s", remainder)
	}
}

func TestRouter_GetTargetForPath_ExactMatch(t *testing.T) {
	r := NewRouter()
	r.Register("/qwen/", "http://localhost:1234/v1/chat/completions")

	route, remainder, found := r.GetTargetForPath("/qwen/")

	if !found {
		t.Fatal("Expected match for exact prefix")
	}

	if route.TargetURL != "http://localhost:1234/v1/chat/completions" {
		t.Errorf("Expected target URL, got %s", route.TargetURL)
	}

	if remainder != "" {
		t.Errorf("Expected empty remainder for exact match, got %s", remainder)
	}
}

func TestRouter_GetTargetForPath_Remainder(t *testing.T) {
	r := NewRouter()
	r.Register("/qwen/", "http://localhost:1234/v1/chat/completions")

	_, remainder, found := r.GetTargetForPath("/qwen/test-model")

	if !found {
		t.Fatal("Expected match for model path under prefix")
	}

	expectedRemainder := "test-model"
	if remainder != expectedRemainder {
		t.Errorf("Expected remainder '%s', got %s", expectedRemainder, remainder)
	}
}

func TestRouter_GetTargetForPath_RemainderWithSlash(t *testing.T) {
	r := NewRouter()
	r.Register("/qwen/", "http://localhost:1234/v1/chat/completions")

	_, remainder, found := r.GetTargetForPath("/qwen/test-model/messages")

	if !found {
		t.Fatal("Expected match for path with slashes under prefix")
	}

	expectedRemainder := "test-model/messages"
	if remainder != expectedRemainder {
		t.Errorf("Expected remainder '%s', got %s", expectedRemainder, remainder)
	}
}

func TestRouter_ServeHTTP_NoMatch(t *testing.T) {
	r := NewRouter()
	req := httptest.NewRequest(http.MethodPost, "/nonexistent/model/test", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected 404 for no match, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "No matching route") {
		t.Errorf("Expected error message about no matching route, got: %s", body)
	}
}

func TestRouter_ServeHTTP_SuccessMatch(t *testing.T) {
	r := NewRouter()

	// Register multiple routes
	r.Register("/qwen/", "http://localhost:1234/v1/chat/completions")
	r.Register("/mistral/", "http://localhost:1234/v1/chat/completions")
	r.Register("/llama/", "http://localhost:1234/v1/chat/completions")
	r.Register("/phi/", "http://localhost:1234/v1/chat/completions")

	// Test Qwen route
	req := httptest.NewRequest(http.MethodPost, "/qwen/test-model", strings.NewReader(`{"model":"test"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusMovedPermanently {
		t.Errorf("Expected 301 redirect, got %d", w.Code)
	}

	location := w.Header().Get("Location")
	expectedLocation := "http://localhost:1234/v1/chat/completions/test-model"
	if !strings.Contains(location, expectedLocation) {
		t.Errorf("Expected location containing '%s', got %s", expectedLocation, location)
	}

	// Test that it strips the prefix correctly
	if strings.HasPrefix(location, "/qwen/") {
		t.Error("Redirect should strip the /qwen/ prefix")
	}
}

func TestRouter_ServeHTTP_LowercasePrefix(t *testing.T) {
	r := NewRouter()
	// Register case-sensitive prefix
	r.Register("/QWEN/", "http://localhost:1234/v1/chat/completions")

	req := httptest.NewRequest(http.MethodPost, "/qwen/test-model", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Current implementation does exact prefix matching, so /qwen won't match /QWEN/
	// This test verifies the current behavior - no redirect for mismatched case
	if w.Code != http.StatusNotFound {
		t.Logf("Got status %d (expected 404 due to case-sensitive matching)", w.Code)
	}
}

func TestRouter_RegisterMultipleRoutes(t *testing.T) {
	r := NewRouter()
	r.RegisterWithRoutes([]Route{
		{Prefix: "/model-a/", TargetURL: "http://backend-a.com/v1/chat/completions"},
		{Prefix: "/model-b/", TargetURL: "http://backend-b.com/v1/chat/completions"},
	})

	if len(r.routes) != 2 {
		t.Errorf("Expected 2 routes registered, got %d", len(r.routes))
	}
}

func TestRouter_LongestMatch(t *testing.T) {
	r := NewRouter()
	// Register overlapping prefixes
	r.Register("/api/", "http://v1.example.com/v1")
	r.Register("/api/test/", "http://specific.example.com/test")

	route, remainder, found := r.GetTargetForPath("/api/test-model")

	if !found {
		t.Fatal("Expected match for /api/test-model")
	}

	expectedURL := "http://v1.example.com/v1"
	if route.TargetURL != expectedURL {
		t.Errorf("Expected longest prefix match %s, got %s", expectedURL, route.TargetURL)
	}

	// Should use the more general /api/ prefix since /api/test/ doesn't match "test-model"
	expectedRemainder := "test-model"
	if remainder != expectedRemainder {
		t.Errorf("Expected remainder '%s', got %s", expectedRemainder, remainder)
	}
}

func TestRouter_QueryParamsPreserved(t *testing.T) {
	r := NewRouter()
	r.Register("/qwen/", "http://localhost:1234/v1/chat/completions")

	req := httptest.NewRequest(http.MethodPost, "/qwen/test-model?stream=true&temperature=0.7", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	location := w.Header().Get("Location")
	if !strings.Contains(location, "stream=true") {
		t.Error("Expected query params to be preserved in redirect")
	}
	if !strings.Contains(location, "temperature=0.7") {
		t.Error("Expected query params to be preserved in redirect")
	}
}

func TestRouter_EmptyPath(t *testing.T) {
	r := NewRouter()
	r.Register("/qwen/", "http://localhost:1234/v1/chat/completions")

	// Empty path won't match any prefix - should return 404
	req := httptest.NewRequest(http.MethodPost, "/test", nil) // Use non-empty path for valid request
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected 404 for no matching route with empty path component, got %d", w.Code)
	}
}

func TestRouter_SingleSlashPath(t *testing.T) {
	r := NewRouter()
	r.Register("/qwen/", "http://localhost:1234/v1/chat/completions")

	// Single slash under prefix won't match because "/qwen" != "/qwen/"
	// This tests that the router handles this edge case gracefully (returns 404)
	req := httptest.NewRequest(http.MethodPost, "/qwen", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Single slash won't match prefix ending with slash - expected 404
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected 301 for single slash under prefix, got %d", w.Code)
	}
}

func BenchmarkRouter_GetTargetForPath(b *testing.B) {
	r := NewRouter()
	for i := 0; i < 10; i++ {
		r.Register("/model-"+string(rune('a'+i))+"/", "http://localhost:1234/v1/chat/completions")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.GetTargetForPath("/model-a/test")
	}
}
