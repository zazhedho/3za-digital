package dto

type SupportContactResponse struct {
	WhatsApp string `json:"whatsapp"`
	Telegram string `json:"telegram"`
	Email    string `json:"email"`
}
