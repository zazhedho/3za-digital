package boqris

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
)

type Client struct {
	config     Config
	httpClient *http.Client
}

func NewClient(config Config) (*Client, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	return NewClientWithHTTPClient(config, &http.Client{Timeout: config.Timeout})
}

func NewClientWithHTTPClient(config Config, httpClient *http.Client) (*Client, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: config.Timeout}
	}
	return &Client{config: config, httpClient: httpClient}, nil
}

func (c *Client) CreateTransaction(ctx context.Context, req CreateTransactionRequest) (*CreateTransactionResponse, error) {
	if strings.TrimSpace(req.MerchantID) == "" {
		req.MerchantID = c.config.MerchantID
	}
	payload := CreateTransactionRequest{
		MerchantID: strings.TrimSpace(req.MerchantID),
		Amount:     req.Amount,
		InvoiceNo:  strings.TrimSpace(req.InvoiceNo),
	}

	var response CreateTransactionResponse
	if err := c.doJSON(ctx, http.MethodPost, "/api/v1/transactions", payload, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *Client) GetTransaction(ctx context.Context, transactionID string) (*CreateTransactionResponse, error) {
	endpoint := path.Join("/api/v1/transactions", strings.TrimSpace(transactionID))
	var response CreateTransactionResponse
	if err := c.doJSON(ctx, http.MethodGet, endpoint, nil, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *Client) doJSON(ctx context.Context, method, endpoint string, payload interface{}, target interface{}) error {
	requestURL, err := c.buildURL(endpoint)
	if err != nil {
		return err
	}

	var body io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = bytes.NewReader(encoded)
	}

	req, err := http.NewRequestWithContext(ctx, method, requestURL, body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(c.config.APIKey))
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return &APIError{
			HTTPStatus: resp.StatusCode,
			Message:    fmt.Sprintf("boqris returned http status %d", resp.StatusCode),
			Raw:        cloneRaw(responseBody),
		}
	}
	if err := json.Unmarshal(responseBody, target); err != nil {
		return err
	}
	setRaw(target, responseBody)
	return nil
}

func (c *Client) buildURL(endpoint string) (string, error) {
	baseURL, err := url.Parse(strings.TrimRight(c.config.BaseURL, "/"))
	if err != nil {
		return "", err
	}
	baseURL.Path = strings.TrimRight(baseURL.Path, "/") + "/" + strings.TrimLeft(endpoint, "/")
	return baseURL.String(), nil
}

func setRaw(target interface{}, body []byte) {
	switch response := target.(type) {
	case *CreateTransactionResponse:
		response.Raw = cloneRaw(body)
	}
}

func cloneRaw(body []byte) json.RawMessage {
	if len(body) == 0 {
		return nil
	}
	cloned := make([]byte, len(body))
	copy(cloned, body)
	return cloned
}
