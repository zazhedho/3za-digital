package boqris

import (
	"errors"
	"strings"
	"time"

	"3za-digital/utils"
)

const defaultBaseURL = "https://api.boqris.id"

type Config struct {
	BaseURL    string
	APIKey     string
	MerchantID string
	Timeout    time.Duration
}

func LoadConfigFromEnv() Config {
	timeoutSeconds := utils.GetEnv("BOQRIS_TIMEOUT_SECONDS", 20)
	if timeoutSeconds <= 0 {
		timeoutSeconds = 20
	}

	return Config{
		BaseURL:    utils.GetEnv("BOQRIS_BASE_URL", defaultBaseURL),
		APIKey:     utils.GetEnv("BOQRIS_API_KEY", ""),
		MerchantID: utils.GetEnv("BOQRIS_MERCHANT_ID", ""),
		Timeout:    time.Duration(timeoutSeconds) * time.Second,
	}
}

func (c Config) Validate() error {
	if strings.TrimSpace(c.BaseURL) == "" {
		return errors.New("boqris base url is required")
	}
	if strings.TrimSpace(c.APIKey) == "" {
		return errors.New("boqris api key is required")
	}
	if strings.TrimSpace(c.MerchantID) == "" {
		return errors.New("boqris merchant id is required")
	}
	return nil
}
