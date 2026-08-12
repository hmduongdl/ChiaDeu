package main

import (
	"io"
	"net/http/httptest"
	"testing"
)

func TestHealthRoutes(t *testing.T) {
	app := newApp()

	for _, path := range []string{"/api/health", "/api/backend/health"} {
		t.Run(path, func(t *testing.T) {
			response, err := app.Test(httptest.NewRequest("GET", path, nil))
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			defer response.Body.Close()

			if response.StatusCode != 200 {
				t.Fatalf("expected status 200, got %d", response.StatusCode)
			}

			body, err := io.ReadAll(response.Body)
			if err != nil {
				t.Fatalf("read response: %v", err)
			}
			if string(body) != `{"status":"ok"}` {
				t.Fatalf("unexpected response body: %s", body)
			}
		})
	}
}
