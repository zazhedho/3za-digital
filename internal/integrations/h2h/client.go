package h2h

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	config     Config
	httpClient *http.Client
	observer   RequestObserver
}

type RequestObserver func(ctx context.Context, event RequestLog)

type RequestLog struct {
	Endpoint       string
	ProductType    string
	RequestRef     string
	ResponseStatus int
	ResponseBody   json.RawMessage
	DurationMS     int
	ErrorMessage   string
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

	return &Client{
		config:     config,
		httpClient: httpClient,
	}, nil
}

func (c *Client) SetObserver(observer RequestObserver) {
	c.observer = observer
}

func (c *Client) GetBalance(ctx context.Context) (*BalanceResponse, error) {
	var response BalanceResponse
	if err := c.get(ctx, "/balance", nil, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *Client) GetPriceList(ctx context.Context, req PriceListRequest) (*PriceListResponse, error) {
	query := url.Values{}
	if req.Type != "" {
		query.Set("type", req.Type)
	}
	if req.Platform != "" {
		query.Set("platform", req.Platform)
	}
	if req.Category != "" {
		query.Set("category", req.Category)
	}

	var response PriceListResponse
	if err := c.get(ctx, "/pricelist", query, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *Client) GetSMMPriceList(ctx context.Context, platform string) (*PriceListResponse, error) {
	return c.GetPriceList(ctx, PriceListRequest{
		Type:     ProductTypeSMM,
		Platform: platform,
	})
}

func (c *Client) CreateOrder(ctx context.Context, req CreateOrderRequest) (*CreateOrderResponse, error) {
	query := url.Values{}
	if req.Type != "" {
		query.Set("type", req.Type)
	}
	if req.ServiceCode != "" {
		if req.Type == ProductTypeSMM {
			query.Set("service", req.ServiceCode)
		} else {
			query.Set("product", req.ServiceCode)
		}
	}
	if req.Target != "" {
		if req.Type == ProductTypeSMM {
			query.Set("target", req.Target)
		} else {
			query.Set("dest", req.Target)
		}
	}
	if req.Quantity > 0 {
		query.Set("quantity", strconv.Itoa(req.Quantity))
	}
	if req.RefID != "" {
		query.Set("refID", req.RefID)
	}
	for key, value := range req.Metadata {
		if value != "" {
			query.Set(key, value)
		}
	}

	var response CreateOrderResponse
	if err := c.get(ctx, "/", query, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *Client) CreateSMMOrder(ctx context.Context, req CreateOrderRequest) (*CreateOrderResponse, error) {
	req.Type = ProductTypeSMM
	return c.CreateOrder(ctx, req)
}

func (c *Client) GetOrderStatus(ctx context.Context, refID string) (*OrderStatusResponse, error) {
	query := url.Values{}
	query.Set("refID", refID)

	var response OrderStatusResponse
	if err := c.get(ctx, "/status", query, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *Client) get(ctx context.Context, endpoint string, query url.Values, target interface{}) error {
	start := time.Now()
	event := RequestLog{
		Endpoint:    displayEndpoint(endpoint),
		ProductType: query.Get("type"),
		RequestRef:  query.Get("refID"),
	}
	observe := func(err error, statusCode int, body []byte) error {
		event.ResponseStatus = statusCode
		event.ResponseBody = cloneRaw(body)
		event.DurationMS = int(time.Since(start).Milliseconds())
		if err != nil {
			event.ErrorMessage = err.Error()
		}
		if c.observer != nil {
			c.observer(ctx, event)
		}
		return err
	}

	requestURL, err := c.buildURL(endpoint, query)
	if err != nil {
		return observe(err, 0, nil)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return observe(err, 0, nil)
	}

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return observe(err, 0, nil)
	}
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return observe(err, httpResp.StatusCode, nil)
	}

	if httpResp.StatusCode < http.StatusOK || httpResp.StatusCode >= http.StatusMultipleChoices {
		return observe(&APIError{
			HTTPStatus: httpResp.StatusCode,
			Message:    fmt.Sprintf("h2h returned http status %d", httpResp.StatusCode),
			Raw:        cloneRaw(body),
		}, httpResp.StatusCode, body)
	}

	if err := json.Unmarshal(body, target); err != nil {
		return observe(err, httpResp.StatusCode, body)
	}
	setRaw(target, body)

	if apiStatus, message := responseStatus(target); !apiStatus {
		return observe(&APIError{
			HTTPStatus: httpResp.StatusCode,
			Message:    message,
			Raw:        cloneRaw(body),
		}, httpResp.StatusCode, body)
	}

	return observe(nil, httpResp.StatusCode, body)
}

func displayEndpoint(endpoint string) string {
	if endpoint == "" || endpoint == "/" {
		return "/"
	}
	return "/" + strings.TrimLeft(endpoint, "/")
}

func (c *Client) buildURL(endpoint string, query url.Values) (string, error) {
	baseURL, err := url.Parse(strings.TrimRight(c.config.BaseURL, "/"))
	if err != nil {
		return "", err
	}

	if endpoint != "" && endpoint != "/" {
		baseURL.Path = strings.TrimRight(baseURL.Path, "/") + "/" + strings.TrimLeft(endpoint, "/")
	}

	values := baseURL.Query()
	values.Set("memberID", c.config.MemberID)
	values.Set("pin", c.config.PIN)
	values.Set("password", c.config.Password)
	for key, value := range query {
		for _, item := range value {
			values.Add(key, item)
		}
	}
	baseURL.RawQuery = values.Encode()

	return baseURL.String(), nil
}

func responseStatus(target interface{}) (bool, string) {
	switch response := target.(type) {
	case *BalanceResponse:
		return response.Status, response.Message
	case *PriceListResponse:
		return response.Status, response.Message
	case *CreateOrderResponse:
		return response.Status, response.Message
	case *OrderStatusResponse:
		return response.Status, response.Message
	default:
		return true, ""
	}
}

func setRaw(target interface{}, body []byte) {
	raw := cloneRaw(body)
	switch response := target.(type) {
	case *BalanceResponse:
		response.Raw = raw
	case *PriceListResponse:
		response.Raw = raw
		for index := range response.Services {
			serviceBody, err := json.Marshal(response.Services[index])
			if err == nil {
				response.Services[index].Raw = cloneRaw(serviceBody)
			}
		}
	case *CreateOrderResponse:
		response.Raw = raw
	case *OrderStatusResponse:
		response.Raw = raw
	}
}

func cloneRaw(body []byte) json.RawMessage {
	raw := make([]byte, len(body))
	copy(raw, body)
	return raw
}
