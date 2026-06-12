package httpjson

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type Request struct {
	Client   *http.Client
	BaseURL  string
	Method   string
	Endpoint string
	Query    url.Values
	Payload  interface{}
	Target   interface{}
	Headers  map[string]string
}

type Response struct {
	StatusCode int
	Body       json.RawMessage
}

func Do(ctx context.Context, req Request) (Response, error) {
	requestURL, err := BuildURL(req.BaseURL, req.Endpoint, req.Query)
	if err != nil {
		return Response{}, err
	}

	var body io.Reader
	if req.Payload != nil {
		encoded, err := json.Marshal(req.Payload)
		if err != nil {
			return Response{}, err
		}
		body = bytes.NewReader(encoded)
	}

	httpReq, err := http.NewRequestWithContext(ctx, req.Method, requestURL, body)
	if err != nil {
		return Response{}, err
	}
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}
	if req.Payload != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	client := req.Client
	if client == nil {
		client = http.DefaultClient
	}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return Response{}, err
	}
	defer httpResp.Body.Close()

	responseBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return Response{}, err
	}
	response := Response{
		StatusCode: httpResp.StatusCode,
		Body:       CloneRaw(responseBody),
	}
	if IsSuccess(httpResp.StatusCode) && req.Target != nil {
		if err := json.Unmarshal(responseBody, req.Target); err != nil {
			return response, err
		}
	}
	return response, nil
}

func BuildURL(base string, endpoint string, query url.Values) (string, error) {
	baseURL, err := url.Parse(strings.TrimRight(base, "/"))
	if err != nil {
		return "", err
	}
	baseURL.Path = strings.TrimRight(baseURL.Path, "/") + "/" + strings.TrimLeft(endpoint, "/")
	if query != nil {
		baseURL.RawQuery = query.Encode()
	}
	return baseURL.String(), nil
}

func IsSuccess(statusCode int) bool {
	return statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices
}

func CloneRaw(body []byte) json.RawMessage {
	if len(body) == 0 {
		return nil
	}
	cloned := make([]byte, len(body))
	copy(cloned, body)
	return cloned
}
