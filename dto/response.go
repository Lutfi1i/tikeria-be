package dto

// WebResponse adalah format respons standar JSON untuk seluruh API
type WebResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

// WebResponses alias untuk kompatibilitas
type WebResponses = WebResponse
