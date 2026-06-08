package repositoryprovider

import (
	"encoding/json"
	"strings"
)

func sanitizeAPILogResponseBody(endpoint string, body json.RawMessage) json.RawMessage {
	if strings.EqualFold(strings.Trim(endpoint, "/"), "pricelist") {
		return json.RawMessage(`{"omitted":true,"reason":"pricelist response body omitted to reduce provider_api_logs size"}`)
	}
	if len(body) == 0 {
		return json.RawMessage(`{}`)
	}
	return body
}
