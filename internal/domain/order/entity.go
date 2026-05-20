package domainorder

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

const (
	StatusPending    = "pending"
	StatusProcessing = "processing"
	StatusCompleted  = "completed"
	StatusPartial    = "partial"
	StatusFailed     = "failed"
	StatusCancelled  = "cancelled" //nolint:misspell
)

func (Order) TableName() string {
	return "orders"
}

type Order struct {
	Id                string          `json:"id" gorm:"column:id;primaryKey"`
	Provider          string          `json:"provider" gorm:"column:provider"`
	ProductType       string          `json:"product_type" gorm:"column:product_type"`
	RefID             string          `json:"ref_id" gorm:"column:ref_id"`
	ServiceID         *string         `json:"service_id,omitempty" gorm:"column:service_id"`
	ProviderServiceID string          `json:"provider_service_id" gorm:"column:provider_service_id"`
	Target            string          `json:"target,omitempty" gorm:"column:target"`
	Quantity          *int64          `json:"quantity,omitempty" gorm:"column:quantity"`
	CustomerNo        string          `json:"customer_no,omitempty" gorm:"column:customer_no"`
	CustomerName      string          `json:"customer_name,omitempty" gorm:"column:customer_name"`
	Status            string          `json:"status" gorm:"column:status"`
	Amount            string          `json:"amount" gorm:"column:amount"`
	ProviderCharge    string          `json:"provider_charge" gorm:"column:provider_charge"`
	Profit            string          `json:"profit" gorm:"column:profit"`
	StartCount        *int64          `json:"start_count,omitempty" gorm:"column:start_count"`
	Remains           *int64          `json:"remains,omitempty" gorm:"column:remains"`
	Metadata          json.RawMessage `json:"metadata" gorm:"column:metadata;type:jsonb"`
	ProviderResponse  json.RawMessage `json:"provider_response" gorm:"column:provider_response;type:jsonb"`
	CreatedBy         *string         `json:"created_by,omitempty" gorm:"column:created_by"`
	CreatedAt         time.Time       `json:"created_at" gorm:"column:created_at"`
	UpdatedAt         *time.Time      `json:"updated_at,omitempty" gorm:"column:updated_at"`
	DeletedAt         gorm.DeletedAt  `json:"-" gorm:"column:deleted_at"`
}

func (OrderStatusLog) TableName() string {
	return "order_status_logs"
}

type OrderStatusLog struct {
	Id               string          `json:"id" gorm:"column:id;primaryKey"`
	OrderID          string          `json:"order_id" gorm:"column:order_id"`
	OldStatus        string          `json:"old_status,omitempty" gorm:"column:old_status"`
	NewStatus        string          `json:"new_status" gorm:"column:new_status"`
	ProviderStatus   string          `json:"provider_status,omitempty" gorm:"column:provider_status"`
	ProviderResponse json.RawMessage `json:"provider_response" gorm:"column:provider_response;type:jsonb"`
	CreatedAt        time.Time       `json:"created_at" gorm:"column:created_at"`
}
