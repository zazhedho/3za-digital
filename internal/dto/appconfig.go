package dto

type UpdateAppConfig struct {
	Value    string `json:"value"`
	IsActive *bool  `json:"is_active" binding:"omitempty"`
}
