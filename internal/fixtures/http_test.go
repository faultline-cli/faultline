package fixtures

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func newHandlerClient(handler http.Handler) *http.Client {
	return &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, req)
			return recorder.Result(), nil
		}),
	}
}

func TestGetJSONUsesConfiguredAcceptHeader(t *testing.T) {
	client := newHandlerClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Header.Get("Accept"), "application/json"; got != want {
			t.Fatalf("expected Accept header %q, got %q", want, got)
		}
		fmt.Fprint(w, `{"ok":true}`)
	}))

	var payload struct {
		OK bool `json:"ok"`
	}
	err := getJSON(context.Background(), client, "https://fixtures.test/example.json", &payload, jsonRequestOptions{
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
	client := newHandlerClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, `{"message":"forbidden"}`)
	}))

	var payload map[string]any
	err := getJSONOptional(context.Background(), client, "https://fixtures.test/optional.json", &payload, jsonRequestOptions{
		AcceptHeader:        "application/json",
		OptionalStatusCodes: []int{http.StatusForbidden},
	})
	if err != nil {
		t.Fatalf("getJSONOptional: %v", err)
	}
}
