package qrisly

import "encoding/json"

type GenerateQRISRequest struct {
	QRISID       string
	Amount       int64
	OutputType   string
	UniqueAmount bool
}

type GenerateQRISResponse struct {
	Success bool             `json:"success"`
	Message string           `json:"message"`
	Data    GenerateQRISData `json:"data"`
	Meta    Meta             `json:"meta"`
	Raw     json.RawMessage  `json:"-"`
}

type GenerateQRISData struct {
	HistoryID      FlexibleString `json:"history_id"`
	QRISID         FlexibleString `json:"qris_id"`
	QRISString     string         `json:"qris_string"`
	QRISImage      string         `json:"qris_image"`
	QRISImageURL   string         `json:"qris_image_url"`
	Image          string         `json:"image"`
	ImageURL       string         `json:"image_url"`
	OriginalAmount FlexibleInt64  `json:"original_amount"`
	FinalAmount    FlexibleInt64  `json:"final_amount"`
	Amount         FlexibleInt64  `json:"amount"`
	PaymentStatus  string         `json:"payment_status"`
	ExpiryTime     string         `json:"expiry_time"`
	ExpiredAt      string         `json:"expired_at"`
	ExpiresAt      string         `json:"expires_at"`
	MerchantName   string         `json:"merchant_name"`
	Name           string         `json:"name"`
}

func (d GenerateQRISData) PayableAmount() int64 {
	if d.FinalAmount.Int64() > 0 {
		return d.FinalAmount.Int64()
	}
	if d.Amount.Int64() > 0 {
		return d.Amount.Int64()
	}
	return d.OriginalAmount.Int64()
}

func (d GenerateQRISData) ImageValue() string {
	if d.QRISImageURL != "" {
		return d.QRISImageURL
	}
	if d.ImageURL != "" {
		return d.ImageURL
	}
	if d.QRISImage != "" {
		return d.QRISImage
	}
	return d.Image
}

func (d GenerateQRISData) ExpiryValue() string {
	if d.ExpiryTime != "" {
		return d.ExpiryTime
	}
	if d.ExpiresAt != "" {
		return d.ExpiresAt
	}
	return d.ExpiredAt
}

func (d GenerateQRISData) MerchantValue() string {
	if d.MerchantName != "" {
		return d.MerchantName
	}
	return d.Name
}

type PaymentStatusResponse struct {
	Success bool              `json:"success"`
	Message string            `json:"message"`
	Meta    Meta              `json:"meta"`
	Data    PaymentStatusData `json:"data"`
	Raw     json.RawMessage   `json:"-"`
}

type PaymentStatusData struct {
	HistoryID     FlexibleString `json:"history_id"`
	PaymentStatus string         `json:"payment_status"`
	Status        string         `json:"status"`
	Amount        FlexibleInt64  `json:"amount"`
	Name          string         `json:"name"`
	PaidAt        FlexibleString `json:"paid_at"`
	CreatedAt     string         `json:"created_at"`
	UpdatedAt     string         `json:"updated_at"`
}

type Meta struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
	Status  string `json:"status"`
}

type APIError struct {
	HTTPStatus int
	Message    string
	Raw        json.RawMessage
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return "qrisly api error"
}
