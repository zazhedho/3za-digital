package qrisly

import (
	"errors"
	"strings"
	"time"

	"3za-digital/utils"
)

const defaultBaseURL = "https://api-sandbox.collaborator.komerce.id/user"

type Config struct {
	BaseURL    string
	APIKey     string
	QRISID     string
	OutputType string
	Timeout    time.Duration
}

func LoadConfigFromEnv() Config {
	timeoutSeconds := utils.GetEnv("QRISLY_TIMEOUT_SECONDS", 20)
	if timeoutSeconds <= 0 {
		timeoutSeconds = 20
	}

	outputType := strings.TrimSpace(utils.GetEnv("QRISLY_OUTPUT_TYPE", "image"))
	if outputType == "" {
		outputType = "image"
	}

	return Config{
		BaseURL:    utils.GetEnv("QRISLY_BASE_URL", defaultBaseURL),
		APIKey:     utils.GetEnv("QRISLY_API_KEY", ""),
		QRISID:     utils.GetEnv("QRISLY_QRIS_ID", ""),
		OutputType: outputType,
		Timeout:    time.Duration(timeoutSeconds) * time.Second,
	}
}

func (c Config) Validate() error {
	if strings.TrimSpace(c.BaseURL) == "" {
		return errors.New("qrisly base url is required")
	}
	if strings.TrimSpace(c.APIKey) == "" {
		return errors.New("qrisly api key is required")
	}
	if strings.TrimSpace(c.QRISID) == "" {
		return errors.New("qrisly qris id is required")
	}
	switch strings.TrimSpace(c.OutputType) {
	case "string", "image":
		return nil
	default:
		return errors.New("qrisly output type must be string or image")
	}
}
