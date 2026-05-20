package money

import "testing"

func TestParseAndFormatCents(t *testing.T) {
	tests := map[string]string{
		"0":       "0.00",
		"1":       "1.00",
		"1.2":     "1.20",
		"1.23":    "1.23",
		"1.234":   "1.23",
		"1.235":   "1.24",
		"999.999": "1000.00",
	}

	for input, want := range tests {
		cents, err := ParseCents(input)
		if err != nil {
			t.Fatalf("ParseCents(%q) returned error: %v", input, err)
		}
		if got := FormatCents(cents); got != want {
			t.Fatalf("FormatCents(ParseCents(%q)) = %q, want %q", input, got, want)
		}
	}
}

func TestMarkupAmount(t *testing.T) {
	provider, amount, profit, err := MarkupAmount("10000", "7.5")
	if err != nil {
		t.Fatalf("MarkupAmount returned error: %v", err)
	}
	if provider != "10000.00" || amount != "10750.00" || profit != "750.00" {
		t.Fatalf("unexpected markup result: provider=%s amount=%s profit=%s", provider, amount, profit)
	}
}
