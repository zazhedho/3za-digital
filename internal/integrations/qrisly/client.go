package qrisly

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
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

func (c *Client) GenerateQRIS(ctx context.Context, req GenerateQRISRequest) (*GenerateQRISResponse, error) {
	if strings.TrimSpace(req.QRISID) == "" {
		req.QRISID = c.config.QRISID
	}
	if strings.TrimSpace(req.OutputType) == "" {
		req.OutputType = c.config.OutputType
	}

	payload := map[string]interface{}{
		"qris_id":       qrisIDValue(req.QRISID),
		"amount":        req.Amount,
		"output_type":   req.OutputType,
		"unique_amount": req.UniqueAmount,
	}
	var response GenerateQRISResponse
	if err := c.doJSON(ctx, http.MethodPost, "/api/v1/qrisly/generate-qris", payload, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *Client) GetPaymentStatus(ctx context.Context, historyID string) (*PaymentStatusResponse, error) {
	endpoint := "/api/v1/qrisly/payment-status/" + url.PathEscape(strings.TrimSpace(historyID))
	var response PaymentStatusResponse
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
	req.Header.Set("X-API-Key", strings.TrimSpace(c.config.APIKey))
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
			Message:    fmt.Sprintf("qrisly returned http status %d", resp.StatusCode),
			Raw:        cloneRaw(responseBody),
		}
	}
	if err := json.Unmarshal(responseBody, target); err != nil {
		return err
	}
	setRaw(target, responseBody)
	if ok, message := responseOK(target); !ok {
		return &APIError{
			HTTPStatus: resp.StatusCode,
			Message:    message,
			Raw:        cloneRaw(responseBody),
		}
	}
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

func qrisIDValue(value string) interface{} {
	value = strings.TrimSpace(value)
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err == nil {
		return parsed
	}
	return value
}

func responseOK(target interface{}) (bool, string) {
	switch response := target.(type) {
	case *GenerateQRISResponse:
		if response.Meta.Status == "error" {
			return false, response.Meta.Message
		}
		if !response.Success && response.Message != "" {
			return false, response.Message
		}
	case *PaymentStatusResponse:
		if response.Meta.Status == "error" {
			return false, response.Meta.Message
		}
		if !response.Success && response.Message != "" {
			return false, response.Message
		}
	}
	return true, ""
}

func setRaw(target interface{}, body []byte) {
	raw := cloneRaw(body)
	switch response := target.(type) {
	case *GenerateQRISResponse:
		response.Raw = raw
	case *PaymentStatusResponse:
		response.Raw = raw
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
