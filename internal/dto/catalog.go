package dto

type SyncCatalogRequest struct {
	Platform string `json:"platform" form:"platform"`
	Category string `json:"category" form:"category"`
}

type SyncCatalogResponse struct {
	Provider    string `json:"provider"`
	ProductType string `json:"product_type"`
	Total       int    `json:"total"`
	Synced      int    `json:"synced"`
}
