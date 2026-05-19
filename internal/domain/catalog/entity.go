package domaincatalog

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

const (
	ProviderH2H = "h2h"

	ProductTypeSMM     = "smm"
	ProductTypePulsa   = "pulsa"
	ProductTypePPOB    = "ppob"
	ProductTypeGame    = "game"
	ProductTypeEWallet = "ewallet"
)

func (ProviderService) TableName() string {
	return "provider_services"
}

type ProviderService struct {
	Id                string          `json:"id" gorm:"column:id;primaryKey"`
	Provider          string          `json:"provider" gorm:"column:provider"`
	ProductType       string          `json:"product_type" gorm:"column:product_type"`
	ProviderServiceID string          `json:"provider_service_id" gorm:"column:provider_service_id"`
	Name              string          `json:"name" gorm:"column:name"`
	Category          string          `json:"category,omitempty" gorm:"column:category"`
	Brand             string          `json:"brand,omitempty" gorm:"column:brand"`
	Platform          string          `json:"platform,omitempty" gorm:"column:platform"`
	MinQuantity       *int64          `json:"min_quantity,omitempty" gorm:"column:min_quantity"`
	MaxQuantity       *int64          `json:"max_quantity,omitempty" gorm:"column:max_quantity"`
	Price             string          `json:"price" gorm:"column:price"`
	Metadata          json.RawMessage `json:"metadata" gorm:"column:metadata;type:jsonb"`
	RawResponse       json.RawMessage `json:"raw_response" gorm:"column:raw_response;type:jsonb"`
	IsActive          bool            `json:"is_active" gorm:"column:is_active"`
	SyncedAt          *time.Time      `json:"synced_at,omitempty" gorm:"column:synced_at"`
	CreatedAt         time.Time       `json:"created_at" gorm:"column:created_at"`
	UpdatedAt         *time.Time      `json:"updated_at,omitempty" gorm:"column:updated_at"`
	DeletedAt         gorm.DeletedAt  `json:"-" gorm:"column:deleted_at"`
}
