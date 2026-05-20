package h2h

import "encoding/json"

const (
	ProductTypeSMM     = "smm"
	ProductTypePulsa   = "pulsa"
	ProductTypePPOB    = "ppob"
	ProductTypeGame    = "game"
	ProductTypeEWallet = "ewallet"
)

type PriceListRequest struct {
	Type     string
	Platform string
	Category string
}

type CreateOrderRequest struct {
	Type        string
	ServiceCode string
	Target      string
	Quantity    int
	RefID       string
	Metadata    map[string]string
}

type BalanceResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Balance FlexibleNumber
	Data    BalanceData     `json:"data"`
	Raw     json.RawMessage `json:"-"`
}

func (r *BalanceResponse) UnmarshalJSON(data []byte) error {
	type wireBalanceResponse struct {
		Status  bool           `json:"status"`
		Message string         `json:"message"`
		Balance FlexibleNumber `json:"balance"`
		Data    BalanceData    `json:"data"`
	}

	var wire wireBalanceResponse
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}

	r.Status = wire.Status
	r.Message = wire.Message
	r.Data = wire.Data
	if wire.Balance.String() != "" {
		r.Balance = wire.Balance
	} else {
		r.Balance = wire.Data.Balance
	}
	return nil
}

type BalanceData struct {
	Balance          FlexibleNumber `json:"balance"`
	BalanceFormatted string         `json:"balance_formatted"`
}

type PriceListResponse struct {
	Status   bool            `json:"status"`
	Message  string          `json:"message"`
	MemberID string          `json:"memberID"`
	Total    int             `json:"total"`
	Services []Service       `json:"data"`
	Raw      json.RawMessage `json:"-"`
}

type Service struct {
	Code        FlexibleString  `json:"code"`
	ID          FlexibleString  `json:"id"`
	Name        string          `json:"name"`
	Category    string          `json:"category"`
	Brand       string          `json:"brand"`
	Platform    string          `json:"platform"`
	Type        string          `json:"type"`
	Price       FlexibleNumber  `json:"price"`
	PricePer1K  FlexibleNumber  `json:"price_per_1k"`
	MinQuantity FlexibleInt     `json:"min"`
	MaxQuantity FlexibleInt     `json:"max"`
	Status      FlexibleString  `json:"status"`
	Raw         json.RawMessage `json:"-"`
}

func (s Service) ProviderServiceID() string {
	if s.Code.String() != "" {
		return s.Code.String()
	}
	return s.ID.String()
}

type CreateOrderResponse struct {
	Status          bool            `json:"status"`
	Message         string          `json:"message"`
	RefID           string          `json:"refID"`
	OrderID         string          `json:"order_id"`
	ProviderOrderID string          `json:"id"`
	ProviderStatus  string          `json:"status_order"`
	Charge          FlexibleNumber  `json:"charge"`
	StartCount      FlexibleNumber  `json:"start_count"`
	Remains         FlexibleNumber  `json:"remains"`
	Raw             json.RawMessage `json:"-"`
}

type OrderStatusResponse struct {
	Status         bool            `json:"status"`
	Message        string          `json:"message"`
	RefID          string          `json:"refID"`
	OrderID        string          `json:"order_id"`
	ProviderStatus string          `json:"status_order"`
	Charge         FlexibleNumber  `json:"charge"`
	StartCount     FlexibleNumber  `json:"start_count"`
	Remains        FlexibleNumber  `json:"remains"`
	Raw            json.RawMessage `json:"-"`
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
	return "h2h api error"
}
