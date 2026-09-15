package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetClientIPUsesConnectionPeer(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.RemoteAddr = "192.0.2.10:12345"
	request.Header.Set("X-Forwarded-For", "198.51.100.20")
	request.Header.Set("X-Real-IP", "203.0.113.30")

	if got := getClientIP(request); got != "192.0.2.10" {
		t.Fatalf("expected connection peer IP, got %q", got)
	}
}

func TestRoutesRejectUnsupportedMethods(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/api/data", nil)
	response := httptest.NewRecorder()

	routes().ServeHTTP(response, request)

	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, response.Code)
	}
}
