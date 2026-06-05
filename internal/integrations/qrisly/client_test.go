package qrisly

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestGenerateQRISUsesConfiguredAuthAndParsesResponse(t *testing.T) {
	var requestBody string
	client, err := NewClientWithHTTPClient(Config{
		BaseURL:    "https://example.test/user",
		APIKey:     "secret-key",
		QRISID:     "18",
		OutputType: "image",
		Timeout:    time.Second,
	}, &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.String() != "https://example.test/user/api/v1/qrisly/generate-qris" {
			t.Fatalf("unexpected url %s", req.URL.String())
		}
		if req.Header.Get("X-API-Key") != "secret-key" {
			t.Fatalf("unexpected api key %q", req.Header.Get("X-API-Key"))
		}
		body, _ := io.ReadAll(req.Body)
		requestBody = string(body)
		return &http.Response{
			StatusCode: http.StatusOK,
			Body: io.NopCloser(bytes.NewBufferString(`{
				"success": true,
				"message": "QRIS successfully generated",
				"data": {
					"history_id": 1778,
					"qris_image_url": "https://qris.test/image.png",
					"original_amount": 105000,
					"final_amount": 105003,
					"payment_status": "unpaid",
					"expiry_time": "2026-03-03 14:52:52"
				}
			}`)),
			Header: make(http.Header),
		}, nil
	})})
	if err != nil {
		t.Fatalf("NewClientWithHTTPClient returned error: %v", err)
	}

	response, err := client.GenerateQRIS(context.Background(), GenerateQRISRequest{Amount: 105000, UniqueAmount: true})
	if err != nil {
		t.Fatalf("GenerateQRIS returned error: %v", err)
	}
	if !strings.Contains(requestBody, `"qris_id":18`) {
		t.Fatalf("expected numeric qris id body, got %s", requestBody)
	}
	if response.Data.HistoryID.String() != "1778" {
		t.Fatalf("expected history id 1778, got %q", response.Data.HistoryID.String())
	}
	if response.Data.PayableAmount() != 105003 {
		t.Fatalf("expected payable 105003, got %d", response.Data.PayableAmount())
	}
	if response.Data.ImageValue() != "https://qris.test/image.png" {
		t.Fatalf("expected image url, got %q", response.Data.ImageValue())
	}
}

func TestGenerateQRISReturnsAPIErrorOnSemanticFailure(t *testing.T) {
	client, err := NewClientWithHTTPClient(Config{
		BaseURL:    "https://example.test/user",
		APIKey:     "secret-key",
		QRISID:     "qris-1",
		OutputType: "string",
		Timeout:    time.Second,
	}, &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(`{"success":false,"message":"QRIS not found"}`)),
			Header:     make(http.Header),
		}, nil
	})})
	if err != nil {
		t.Fatalf("NewClientWithHTTPClient returned error: %v", err)
	}

	_, err = client.GenerateQRIS(context.Background(), GenerateQRISRequest{Amount: 1000, UniqueAmount: true})
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T %v", err, err)
	}
	if apiErr.Message != "QRIS not found" {
		t.Fatalf("expected QRIS not found, got %q", apiErr.Message)
	}
}
