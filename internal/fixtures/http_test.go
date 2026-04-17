package fixtures

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetJSONUsesConfiguredAcceptHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Header.Get("Accept"), "application/json"; got != want {
			t.Fatalf("expected Accept header %q, got %q", want, got)
		}
		fmt.Fprint(w, `{"ok":true}`)
	}))
	defer server.Close()

	var payload struct {
		OK bool `json:"ok"`
	}
	err := getJSON(context.Background(), server.Client(), server.URL, &payload, jsonRequestOptions{
		AcceptHeader: "application/json",
	})
	if err != nil {
		t.Fatalf("getJSON: %v", err)
	}
	if !payload.OK {
		t.Fatal("expected decoded payload")
	}
}

func TestGetJSONOptionalIgnoresConfiguredStatusCodes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, `{"message":"forbidden"}`)
	}))
	defer server.Close()

	var payload map[string]any
	err := getJSONOptional(context.Background(), server.Client(), server.URL, &payload, jsonRequestOptions{
		AcceptHeader:        "application/json",
		OptionalStatusCodes: []int{http.StatusForbidden},
	})
	if err != nil {
		t.Fatalf("getJSONOptional: %v", err)
	}
}
