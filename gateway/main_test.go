package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	healthHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var body HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body.Service != "gateway" {
		t.Errorf("expected service 'gateway', got '%s'", body.Service)
	}
	if body.Status != "healthy" {
		t.Errorf("expected status 'healthy', got '%s'", body.Status)
	}
	if body.Timestamp == "" {
		t.Error("expected non-empty timestamp")
	}
}

func TestProxyHealthHandler_Success(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"service": "mock",
			"status":  "healthy",
		})
	}))
	defer backend.Close()

	target := ProxyTarget{Name: "mock", URL: backend.URL}
	handler := proxyHealthHandler(target)

	req := httptest.NewRequest(http.MethodGet, "/api/mock/health", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	if body["status"] != "healthy" {
		t.Errorf("expected status 'healthy', got '%v'", body["status"])
	}
}

func TestProxyHealthHandler_BackendDown(t *testing.T) {
	target := ProxyTarget{Name: "down-service", URL: "http://localhost:1"}
	handler := proxyHealthHandler(target)

	req := httptest.NewRequest(http.MethodGet, "/api/down/health", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadGateway {
		t.Errorf("expected status 502, got %d", resp.StatusCode)
	}
}

func TestStatusHandler(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"service": "mock",
			"status":  "healthy",
		})
	}))
	defer backend.Close()

	targets := []ProxyTarget{
		{Name: "mock", URL: backend.URL},
	}
	handler := statusHandler(targets)

	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	if body["gateway"] == nil {
		t.Error("expected gateway in status response")
	}
	if body["mock"] == nil {
		t.Error("expected mock service in status response")
	}
}

func TestGetEnv(t *testing.T) {
	val := getEnv("NONEXISTENT_VAR_12345", "default")
	if val != "default" {
		t.Errorf("expected 'default', got '%s'", val)
	}

	t.Setenv("TEST_GETENV_VAR", "custom")
	val = getEnv("TEST_GETENV_VAR", "default")
	if val != "custom" {
		t.Errorf("expected 'custom', got '%s'", val)
	}
}
