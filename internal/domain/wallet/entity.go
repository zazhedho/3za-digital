package domainwallet

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

const (
	TransactionTypeDeposit     = "deposit"
	TransactionTypeDebitOrder  = "debit_order"
	TransactionTypeRefundOrder = "refund_order"
	TransactionTypeAdjustment  = "adjustment"

	DirectionCredit = "credit"
	DirectionDebit  = "debit"

	DepositStatusPending   = "pending"
	DepositStatusPaid      = "paid"
	DepositStatusExpired   = "expired"
	DepositStatusFailed    = "failed"
	DepositStatusCancelled = "cancelled" //nolint:misspell

	DepositMethodManualAdmin    = "manual_admin"
	DepositMethodPaymentGateway = "payment_gateway"
)

func (Wallet) TableName() string {
	return "wallets"
}

type Wallet struct {
	Id            string         `json:"id" gorm:"column:id;primaryKey"`
	UserID        string         `json:"user_id" gorm:"column:user_id"`
	User          *UserSummary   `json:"user,omitempty" gorm:"foreignKey:UserID;references:Id"`
	Balance       string         `json:"balance" gorm:"column:balance"`
	LockedBalance string         `json:"locked_balance" gorm:"column:locked_balance"`
	Currency      string         `json:"currency" gorm:"column:currency"`
	IsActive      bool           `json:"is_active" gorm:"column:is_active"`
	CreatedAt     time.Time      `json:"created_at" gorm:"column:created_at"`
	UpdatedAt     *time.Time     `json:"updated_at,omitempty" gorm:"column:updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"column:deleted_at"`
}

func (WalletTransaction) TableName() string {
	return "wallet_transactions"
}

type WalletTransaction struct {
	Id               string          `json:"id" gorm:"column:id;primaryKey"`
	WalletID         string          `json:"wallet_id" gorm:"column:wallet_id"`
	UserID           string          `json:"user_id" gorm:"column:user_id"`
	OrderID          *string         `json:"order_id,omitempty" gorm:"column:order_id"`
	DepositRequestID *string         `json:"deposit_request_id,omitempty" gorm:"column:deposit_request_id"`
	Type             string          `json:"type" gorm:"column:type"`
	Direction        string          `json:"direction" gorm:"column:direction"`
	Amount           string          `json:"amount" gorm:"column:amount"`
	BalanceBefore    string          `json:"balance_before" gorm:"column:balance_before"`
	BalanceAfter     string          `json:"balance_after" gorm:"column:balance_after"`
	Reference        string          `json:"reference,omitempty" gorm:"column:reference"`
	Description      string          `json:"description,omitempty" gorm:"column:description"`
	Metadata         json.RawMessage `json:"metadata" gorm:"column:metadata;type:jsonb"`
	CreatedBy        *string         `json:"created_by,omitempty" gorm:"column:created_by"`
	CreatedAt        time.Time       `json:"created_at" gorm:"column:created_at"`
}

func (DepositRequest) TableName() string {
	return "deposit_requests"
}

type DepositRequest struct {
	Id               string          `json:"id" gorm:"column:id;primaryKey"`
	UserID           string          `json:"user_id" gorm:"column:user_id"`
	User             *UserSummary    `json:"user,omitempty" gorm:"foreignKey:UserID;references:Id"`
	Amount           string          `json:"amount" gorm:"column:amount"`
	Status           string          `json:"status" gorm:"column:status"`
	Method           string          `json:"method" gorm:"column:method"`
	Provider         string          `json:"provider,omitempty" gorm:"column:provider"`
	PaymentReference string          `json:"payment_reference,omitempty" gorm:"column:payment_reference"`
	PaymentURL       string          `json:"payment_url,omitempty" gorm:"column:payment_url"`
	ExpiredAt        *time.Time      `json:"expired_at,omitempty" gorm:"column:expired_at"`
	PaidAt           *time.Time      `json:"paid_at,omitempty" gorm:"column:paid_at"`
	Metadata         json.RawMessage `json:"metadata" gorm:"column:metadata;type:jsonb"`
	CreatedBy        *string         `json:"created_by,omitempty" gorm:"column:created_by"`
	CreatedAt        time.Time       `json:"created_at" gorm:"column:created_at"`
	UpdatedAt        *time.Time      `json:"updated_at,omitempty" gorm:"column:updated_at"`
	DeletedAt        gorm.DeletedAt  `json:"-" gorm:"column:deleted_at"`
}

func (UserSummary) TableName() string {
	return "users"
}

type UserSummary struct {
	Id        string `json:"id" gorm:"column:id;primaryKey"`
	Name      string `json:"name" gorm:"column:name"`
	Email     string `json:"email,omitempty" gorm:"column:email"`
	Phone     string `json:"phone,omitempty" gorm:"column:phone"`
	Role      string `json:"role,omitempty" gorm:"column:role"`
	AvatarURL string `json:"avatar_url,omitempty" gorm:"column:avatar_url"`
}

func (PaymentGatewayLog) TableName() string {
	return "payment_gateway_logs"
}

type PaymentGatewayLog struct {
	Id               string          `json:"id" gorm:"column:id;primaryKey"`
	Provider         string          `json:"provider" gorm:"column:provider"`
	EventType        string          `json:"event_type,omitempty" gorm:"column:event_type"`
	RequestID        string          `json:"request_id,omitempty" gorm:"column:request_id"`
	PaymentReference string          `json:"payment_reference,omitempty" gorm:"column:payment_reference"`
	DepositRequestID *string         `json:"deposit_request_id,omitempty" gorm:"column:deposit_request_id"`
	Signature        string          `json:"signature,omitempty" gorm:"column:signature"`
	Status           string          `json:"status,omitempty" gorm:"column:status"`
	Payload          json.RawMessage `json:"payload" gorm:"column:payload;type:jsonb"`
	ErrorMessage     string          `json:"error_message,omitempty" gorm:"column:error_message"`
	CreatedAt        time.Time       `json:"created_at" gorm:"column:created_at"`
}
