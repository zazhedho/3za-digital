package qrisly

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
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
	response, err := httpjson.Do(ctx, httpjson.Request{
		Client:   c.httpClient,
		BaseURL:  c.config.BaseURL,
		Method:   method,
		Endpoint: endpoint,
		Payload:  payload,
		Target:   target,
		Headers: map[string]string{
			"X-API-Key": strings.TrimSpace(c.config.APIKey),
		},
	})
	if err != nil {
		return err
	}
	if !httpjson.IsSuccess(response.StatusCode) {
		return &APIError{
			HTTPStatus: response.StatusCode,
			Message:    fmt.Sprintf("qrisly returned http status %d", response.StatusCode),
			Raw:        response.Body,
		}
	}
	setRaw(target, response.Body)
	if ok, message := responseOK(target); !ok {
		return &APIError{
			HTTPStatus: response.StatusCode,
			Message:    message,
			Raw:        response.Body,
		}
	}
	return nil
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
	raw := httpjson.CloneRaw(body)
	switch response := target.(type) {
	case *GenerateQRISResponse:
		response.Raw = raw
	case *PaymentStatusResponse:
		response.Raw = raw
	}
}
