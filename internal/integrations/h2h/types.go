package h2h

import (
	"encoding/json"

	"3za-digital/utils"
)

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
	Balance utils.FlexibleNumber
	Data    BalanceData     `json:"data"`
	Raw     json.RawMessage `json:"-"`
}

func (r *BalanceResponse) UnmarshalJSON(data []byte) error {
	type wireBalanceResponse struct {
		Status  bool                 `json:"status"`
		Message string               `json:"message"`
		Balance utils.FlexibleNumber `json:"balance"`
		Data    BalanceData          `json:"data"`
	}
	return unmarshalWithWire(data, r, &wireBalanceResponse{}, func(target *BalanceResponse, wire *wireBalanceResponse) {
		target.Status = wire.Status
		target.Message = wire.Message
		target.Data = wire.Data
		if wire.Balance.String() != "" {
			target.Balance = wire.Balance
		} else {
			target.Balance = wire.Data.Balance
		}
	})
}

type BalanceData struct {
	Balance          utils.FlexibleNumber `json:"balance"`
	BalanceFormatted string               `json:"balance_formatted"`
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
	Code        utils.FlexibleString `json:"code"`
	ID          utils.FlexibleString `json:"id"`
	Name        string               `json:"name"`
	Category    string               `json:"category"`
	Brand       string               `json:"brand"`
	Platform    string               `json:"platform"`
	Type        string               `json:"type"`
	Price       utils.FlexibleNumber `json:"price"`
	PricePer1K  utils.FlexibleNumber `json:"price_per_1k"`
	MinQuantity utils.FlexibleInt    `json:"min"`
	MaxQuantity utils.FlexibleInt    `json:"max"`
	Status      utils.FlexibleString `json:"status"`
	Raw         json.RawMessage      `json:"-"`
}

func (s Service) ProviderServiceID() string {
	if s.Code.String() != "" {
		return s.Code.String()
	}
	return s.ID.String()
}

type CreateOrderResponse struct {
	Status          bool                 `json:"status"`
	Message         string               `json:"message"`
	RefID           string               `json:"refID"`
	OrderID         string               `json:"order_id"`
	ProviderOrderID string               `json:"id"`
	ProviderStatus  string               `json:"status_order"`
	Charge          utils.FlexibleNumber `json:"charge"`
	StartCount      utils.FlexibleNumber `json:"start_count"`
	Remains         utils.FlexibleNumber `json:"remains"`
	Raw             json.RawMessage      `json:"-"`
}

type OrderStatusResponse struct {
	Status         bool                 `json:"status"`
	Message        string               `json:"message"`
	RefID          string               `json:"refID"`
	OrderID        string               `json:"order_id"`
	ProviderStatus string               `json:"status_order"`
	Charge         utils.FlexibleNumber `json:"charge"`
	StartCount     utils.FlexibleNumber `json:"start_count"`
	Remains        utils.FlexibleNumber `json:"remains"`
	Raw            json.RawMessage      `json:"-"`
}

func (r *OrderStatusResponse) UnmarshalJSON(data []byte) error {
	type alias OrderStatusResponse
	type wireOrderStatusResponse struct {
		alias
		RefIDSnake        string               `json:"ref_id"`
		TransactionStatus string               `json:"transaction_status"`
		StatusLabel       string               `json:"status_label"`
		StatusDescription string               `json:"status_description"`
		Price             utils.FlexibleNumber `json:"price"`
		Data              *struct {
			OrderNumber       string               `json:"order_number"`
			RefID             string               `json:"ref_id"`
			ProviderStatus    string               `json:"status_order"`
			TransactionStatus string               `json:"transaction_status"`
			StatusLabel       string               `json:"status_label"`
			StatusDescription string               `json:"status_description"`
			Price             utils.FlexibleNumber `json:"price"`
			Charge            utils.FlexibleNumber `json:"charge"`
			StartCount        utils.FlexibleNumber `json:"start_count"`
			Remains           utils.FlexibleNumber `json:"remains"`
		} `json:"data"`
	}
	return unmarshalWithWire(data, r, &wireOrderStatusResponse{}, func(target *OrderStatusResponse, decoded *wireOrderStatusResponse) {
		*target = OrderStatusResponse(decoded.alias)
		if target.RefID == "" {
			target.RefID = decoded.RefIDSnake
		}
		if target.ProviderStatus == "" {
			target.ProviderStatus = utils.FirstNonEmptyString(decoded.TransactionStatus, decoded.StatusLabel, decoded.StatusDescription)
		}
		if target.Charge.String() == "" && decoded.Price.String() != "" {
			target.Charge = decoded.Price
		}

		if decoded.Data != nil {
			if target.RefID == "" {
				target.RefID = decoded.Data.RefID
			}
			if target.OrderID == "" {
				target.OrderID = decoded.Data.OrderNumber
			}
			if target.ProviderStatus == "" {
				target.ProviderStatus = utils.FirstNonEmptyString(
					decoded.Data.TransactionStatus,
					decoded.Data.ProviderStatus,
					decoded.Data.StatusLabel,
					decoded.Data.StatusDescription,
				)
			}
			if target.Charge.String() == "" {
				target.Charge = utils.FirstFlexibleNumber(decoded.Data.Charge, decoded.Data.Price)
			}
			if target.StartCount.String() == "" {
				target.StartCount = decoded.Data.StartCount
			}
			if target.Remains.String() == "" {
				target.Remains = decoded.Data.Remains
			}
		}
	})
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
