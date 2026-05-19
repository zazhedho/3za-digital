package domaindashboard

type Summary struct {
	TotalOrders         int64  `json:"total_orders"`
	PendingOrders       int64  `json:"pending_orders"`
	ProcessingOrders    int64  `json:"processing_orders"`
	CompletedOrders     int64  `json:"completed_orders"`
	PartialOrders       int64  `json:"partial_orders"`
	FailedOrders        int64  `json:"failed_orders"`
	CancelledOrders     int64  `json:"cancelled_orders"`
	TotalAmount         string `json:"total_amount"`
	TotalProviderCharge string `json:"total_provider_charge"`
	TotalProfit         string `json:"total_profit"`
}
