package repositoryprovider

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSanitizeAPILogResponseBodyOmitsPricelistBody(t *testing.T) {
	body := json.RawMessage(`{"data":[{"id":1},{"id":2}]}`)

	got := sanitizeAPILogResponseBody("/pricelist", body)

	if strings.Contains(string(got), `"data"`) {
		t.Fatalf("expected pricelist response body to be omitted, got %s", string(got))
	}
	if !strings.Contains(string(got), `"omitted":true`) {
		t.Fatalf("expected omitted marker, got %s", string(got))
	}
}

func TestSanitizeAPILogResponseBodyKeepsNonPricelistBody(t *testing.T) {
	body := json.RawMessage(`{"status":true}`)

	got := sanitizeAPILogResponseBody("/status", body)

	if string(got) != string(body) {
		t.Fatalf("expected response body unchanged, got %s", string(got))
	}
}

func TestSanitizeAPILogResponseBodyDefaultsEmptyBody(t *testing.T) {
	got := sanitizeAPILogResponseBody("/status", nil)

	if string(got) != `{}` {
		t.Fatalf("expected empty json object, got %s", string(got))
	}
}
