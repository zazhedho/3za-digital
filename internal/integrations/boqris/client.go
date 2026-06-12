package boqris

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"strings"

	"3za-digital/internal/integrations/httpjson"
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
	response, err := httpjson.Do(ctx, httpjson.Request{
		Client:   c.httpClient,
		BaseURL:  c.config.BaseURL,
		Method:   method,
		Endpoint: endpoint,
		Payload:  payload,
		Target:   target,
		Headers: map[string]string{
			"Authorization": "Bearer " + strings.TrimSpace(c.config.APIKey),
		},
	})
	if err != nil {
		return err
	}
	if !httpjson.IsSuccess(response.StatusCode) {
		return &APIError{
			HTTPStatus: response.StatusCode,
			Message:    fmt.Sprintf("boqris returned http status %d", response.StatusCode),
			Raw:        response.Body,
		}
	}
	setRaw(target, response.Body)
	return nil
}

func setRaw(target interface{}, body []byte) {
	switch response := target.(type) {
	case *CreateTransactionResponse:
		response.Raw = httpjson.CloneRaw(body)
	}
}
