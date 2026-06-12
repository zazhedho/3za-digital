package httpjson

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestDoSendsJSONHeadersAndDecodesResponse(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.String() != "https://api.test/base/v1/items?status=pending" {
			t.Fatalf("unexpected url %s", req.URL.String())
		}
		if req.Method != http.MethodPost {
			t.Fatalf("unexpected method %s", req.Method)
		}
		if req.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("unexpected authorization %q", req.Header.Get("Authorization"))
		}
		if req.Header.Get("Content-Type") != "application/json" {
			t.Fatalf("unexpected content type %q", req.Header.Get("Content-Type"))
		}
		body, _ := io.ReadAll(req.Body)
		if string(body) != `{"name":"deposit"}` {
			t.Fatalf("unexpected body %s", body)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"id":"trx-1"}`)),
		}, nil
	})}

	var target struct {
		ID string `json:"id"`
	}
	response, err := Do(context.Background(), Request{
		Client:   client,
		BaseURL:  "https://api.test/base",
		Method:   http.MethodPost,
		Endpoint: "/v1/items",
		Query:    url.Values{"status": []string{"pending"}},
		Payload:  map[string]string{"name": "deposit"},
		Target:   &target,
		Headers:  map[string]string{"Authorization": "Bearer token"},
	})
	if err != nil {
		t.Fatalf("Do returned error: %v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status %d", response.StatusCode)
	}
	if target.ID != "trx-1" {
		t.Fatalf("expected decoded target, got %q", target.ID)
	}
	if string(response.Body) != `{"id":"trx-1"}` {
		t.Fatalf("unexpected raw body %s", response.Body)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
