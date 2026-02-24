// Package testutil provides shared helpers for HTTP handler tests.
package testutil

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// NewJSONRequest builds an HTTP request with a JSON-encoded body.
// Pass body=nil to create a request with no body.
func NewJSONRequest(t testing.TB, method, target string, body any) *http.Request {
	t.Helper()
	if body == nil {
		req, err := http.NewRequest(method, target, nil)
		if err != nil {
			t.Fatalf("testutil.NewJSONRequest: %v", err)
		}
		return req
	}
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("testutil.NewJSONRequest: marshal body: %v", err)
	}
	req, err := http.NewRequest(method, target, bytes.NewReader(b))
	if err != nil {
		t.Fatalf("testutil.NewJSONRequest: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	return req
}

// ParseJSON decodes the response body as a JSON object and returns it as a map.
// Fails the test immediately if decoding fails.
func ParseJSON(t testing.TB, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &m); err != nil {
		t.Fatalf("testutil.ParseJSON: body=%q err=%v", w.Body.String(), err)
	}
	return m
}

// AssertStatus fails the test if the response status code does not match want.
func AssertStatus(t testing.TB, w *httptest.ResponseRecorder, want int) {
	t.Helper()
	if w.Code != want {
		t.Errorf("status = %d; want %d; body = %s", w.Code, want, w.Body.String())
	}
}
