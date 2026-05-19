package h2h

import (
	"errors"
	"time"

	"3za-digital/utils"
)

const defaultBaseURL = "https://api.h2h.id/api/trx"

type Config struct {
	BaseURL  string
	MemberID string
	PIN      string
	Password string
	Timeout  time.Duration
}

func LoadConfigFromEnv() Config {
	timeoutSeconds := utils.GetEnv("H2H_TIMEOUT_SECONDS", 20)
	if timeoutSeconds <= 0 {
		timeoutSeconds = 20
	}

	return Config{
		BaseURL:  utils.GetEnv("H2H_BASE_URL", defaultBaseURL),
		MemberID: utils.GetEnv("H2H_MEMBER_ID", ""),
		PIN:      utils.GetEnv("H2H_PIN", ""),
		Password: utils.GetEnv("H2H_PASSWORD", ""),
		Timeout:  time.Duration(timeoutSeconds) * time.Second,
	}
}

func (c Config) Validate() error {
	if c.BaseURL == "" {
		return errors.New("h2h base url is required")
	}
	if c.MemberID == "" {
		return errors.New("h2h member id is required")
	}
	if c.PIN == "" {
		return errors.New("h2h pin is required")
	}
	if c.Password == "" {
		return errors.New("h2h password is required")
	}
	return nil
}
