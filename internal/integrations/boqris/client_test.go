package boqris

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestGetTransactionUsesBearerAuthAndParsesResponse(t *testing.T) {
	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Path != "/api/v1/transactions/ef7cabdb-1e93-4a68-bdfe-dd92894233c3" {
			t.Fatalf("expected status path, got %s", req.URL.Path)
		}
		if req.Header.Get("Authorization") != "Bearer secret-key" {
			t.Fatalf("expected bearer auth, got %q", req.Header.Get("Authorization"))
		}
		body := `{
			"transaction_id": "ef7cabdb-1e93-4a68-bdfe-dd92894233c3",
			"status": "expired",
			"merchant_id": "merchant-1",
			"base_amount": 25000,
			"unique_code": 762,
			"amount": 25762,
			"invoice_no": null,
			"qris_dynamic": "000201010212",
			"qr_url": "https://api.boqris.id/qr/ef7cabdb-1e93-4a68-bdfe-dd92894233c3",
			"expires_at": "2026-06-12 03:58:27",
			"paid_at": null,
			"created_at": "2026-06-12 03:43:27"
		}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(body)),
		}, nil
	})

	client, err := NewClientWithHTTPClient(Config{
		BaseURL:    "https://api.boqris.test",
		APIKey:     "secret-key",
		MerchantID: "merchant-1",
		Timeout:    time.Second,
	}, &http.Client{Transport: transport})
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	response, err := client.GetTransaction(context.Background(), "ef7cabdb-1e93-4a68-bdfe-dd92894233c3")
	if err != nil {
		t.Fatalf("GetTransaction returned error: %v", err)
	}
	if response.Status != "expired" {
		t.Fatalf("expected expired status, got %q", response.Status)
	}
	if response.Amount.Int64() != 25762 {
		t.Fatalf("expected amount 25762, got %d", response.Amount.Int64())
	}
	if len(response.Raw) == 0 {
		t.Fatal("expected raw response")
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
