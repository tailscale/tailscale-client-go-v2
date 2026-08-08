package openapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetch(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("openapi: 3.1.0\n"))
	}))
	defer server.Close()

	data, err := Fetch(context.Background(), server.Client(), server.URL)
	if err != nil {
		t.Fatalf("Fetch returned error: %v", err)
	}

	if got, want := string(data), "openapi: 3.1.0\n"; got != want {
		t.Fatalf("Fetch returned %q, want %q", got, want)
	}
}
