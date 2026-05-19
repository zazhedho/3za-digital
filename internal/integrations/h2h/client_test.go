package h2h

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClientGetBalance(t *testing.T) {
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertAuth(t, r)
		if r.URL.Path != "/balance" {
			t.Fatalf("expected /balance path, got %s", r.URL.Path)
		}

		writeJSON(t, w, map[string]interface{}{
			"status":  true,
			"message": "success",
			"balance": 125000,
		})
	})
	defer server.Close()

	client := newTestClient(t, server.URL)
	response, err := client.GetBalance(context.Background())
	if err != nil {
		t.Fatalf("GetBalance returned error: %v", err)
	}

	if !response.Status {
		t.Fatal("expected status true")
	}
	if response.Balance.String() != "125000" {
		t.Fatalf("expected balance 125000, got %s", response.Balance.String())
	}
	if len(response.Raw) == 0 {
		t.Fatal("expected raw response")
	}
}

func TestClientGetSMMPriceList(t *testing.T) {
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertAuth(t, r)
		if r.URL.Path != "/pricelist" {
			t.Fatalf("expected /pricelist path, got %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("type"); got != ProductTypeSMM {
			t.Fatalf("expected type smm, got %s", got)
		}
		if got := r.URL.Query().Get("platform"); got != "instagram" {
			t.Fatalf("expected platform instagram, got %s", got)
		}

		writeJSON(t, w, map[string]interface{}{
			"status":   true,
			"message":  "success",
			"memberID": "member-1",
			"total":    1,
			"data": []map[string]interface{}{
				{
					"id":       1001,
					"name":     "Instagram Followers",
					"category": "Followers",
					"platform": "instagram",
					"price":    "1000",
					"min":      "10",
					"max":      "10000",
					"status":   1,
				},
			},
		})
	})
	defer server.Close()

	client := newTestClient(t, server.URL)
	response, err := client.GetSMMPriceList(context.Background(), "instagram")
	if err != nil {
		t.Fatalf("GetSMMPriceList returned error: %v", err)
	}

	if response.Total != 1 {
		t.Fatalf("expected total 1, got %d", response.Total)
	}
	if got := response.Services[0].ProviderServiceID(); got != "1001" {
		t.Fatalf("expected provider service id 1001, got %s", got)
	}
}

func TestClientCreateSMMOrder(t *testing.T) {
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertAuth(t, r)
		if r.URL.Path != "/" {
			t.Fatalf("expected / path, got %s", r.URL.Path)
		}
		query := r.URL.Query()
		expected := map[string]string{
			"type":     ProductTypeSMM,
			"service":  "1001",
			"target":   "https://instagram.com/3zadigital",
			"quantity": "100",
			"refID":    "SMM-001",
		}
		for key, value := range expected {
			if got := query.Get(key); got != value {
				t.Fatalf("expected query %s=%s, got %s", key, value, got)
			}
		}

		writeJSON(t, w, map[string]interface{}{
			"status":       true,
			"message":      "success",
			"refID":        "SMM-001",
			"order_id":     "ORD-1",
			"status_order": "pending",
			"charge":       1000,
		})
	})
	defer server.Close()

	client := newTestClient(t, server.URL)
	response, err := client.CreateSMMOrder(context.Background(), CreateOrderRequest{
		ServiceCode: "1001",
		Target:      "https://instagram.com/3zadigital",
		Quantity:    100,
		RefID:       "SMM-001",
	})
	if err != nil {
		t.Fatalf("CreateSMMOrder returned error: %v", err)
	}

	if response.RefID != "SMM-001" {
		t.Fatalf("expected refID SMM-001, got %s", response.RefID)
	}
	if response.ProviderStatus != "pending" {
		t.Fatalf("expected provider status pending, got %s", response.ProviderStatus)
	}
}

func TestClientGetOrderStatus(t *testing.T) {
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertAuth(t, r)
		if r.URL.Path != "/status" {
			t.Fatalf("expected /status path, got %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("refID"); got != "SMM-001" {
			t.Fatalf("expected refID SMM-001, got %s", got)
		}

		writeJSON(t, w, map[string]interface{}{
			"status":       true,
			"message":      "success",
			"refID":        "SMM-001",
			"order_id":     "ORD-1",
			"status_order": "completed",
			"remains":      0,
		})
	})
	defer server.Close()

	client := newTestClient(t, server.URL)
	response, err := client.GetOrderStatus(context.Background(), "SMM-001")
	if err != nil {
		t.Fatalf("GetOrderStatus returned error: %v", err)
	}
	if response.ProviderStatus != "completed" {
		t.Fatalf("expected completed status, got %s", response.ProviderStatus)
	}
}

func TestClientReturnsAPIErrorForFailedPayload(t *testing.T) {
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertAuth(t, r)
		writeJSON(t, w, map[string]interface{}{
			"status":  false,
			"message": "invalid credential",
		})
	})
	defer server.Close()

	client := newTestClient(t, server.URL)
	_, err := client.GetBalance(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.Message != "invalid credential" {
		t.Fatalf("expected invalid credential message, got %s", apiErr.Message)
	}
}

func newTestServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	return httptest.NewServer(handler)
}

func newTestClient(t *testing.T, baseURL string) *Client {
	t.Helper()

	client, err := NewClientWithHTTPClient(Config{
		BaseURL:  baseURL,
		MemberID: "member-1",
		PIN:      "123456",
		Password: "secret",
		Timeout:  time.Second,
	}, http.DefaultClient)
	if err != nil {
		t.Fatalf("NewClientWithHTTPClient returned error: %v", err)
	}
	return client
}

func assertAuth(t *testing.T, r *http.Request) {
	t.Helper()

	query := r.URL.Query()
	expected := map[string]string{
		"memberID": "member-1",
		"pin":      "123456",
		"password": "secret",
	}
	for key, value := range expected {
		if got := query.Get(key); got != value {
			t.Fatalf("expected auth query %s=%s, got %s", key, value, got)
		}
	}
}

func writeJSON(t *testing.T, w http.ResponseWriter, payload interface{}) {
	t.Helper()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		t.Fatalf("failed to write json: %v", err)
	}
}
