package domainprovider

import (
	"encoding/json"
	"time"
)

func (BalanceSnapshot) TableName() string {
	return "provider_balance_snapshots"
}

type BalanceSnapshot struct {
	Id          string          `json:"id" gorm:"column:id;primaryKey"`
	Provider    string          `json:"provider" gorm:"column:provider"`
	Balance     string          `json:"balance" gorm:"column:balance"`
	RawResponse json.RawMessage `json:"raw_response" gorm:"column:raw_response;type:jsonb"`
	CreatedAt   time.Time       `json:"created_at" gorm:"column:created_at"`
}

func (APILog) TableName() string {
	return "provider_api_logs"
}

type APILog struct {
	Id             string          `json:"id" gorm:"column:id;primaryKey"`
	Provider       string          `json:"provider" gorm:"column:provider"`
	ProductType    string          `json:"product_type,omitempty" gorm:"column:product_type"`
	Endpoint       string          `json:"endpoint" gorm:"column:endpoint"`
	RequestRef     string          `json:"request_ref,omitempty" gorm:"column:request_ref"`
	ResponseStatus *int            `json:"response_status,omitempty" gorm:"column:response_status"`
	ResponseBody   json.RawMessage `json:"response_body" gorm:"column:response_body;type:jsonb"`
	DurationMS     *int            `json:"duration_ms,omitempty" gorm:"column:duration_ms"`
	ErrorMessage   string          `json:"error_message,omitempty" gorm:"column:error_message"`
	CreatedAt      time.Time       `json:"created_at" gorm:"column:created_at"`
}
