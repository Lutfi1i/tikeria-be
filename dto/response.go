package dto

// WebResponse adalah format respons standar JSON untuk seluruh API
type WebResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	ErrorCode string      `json:"error_code,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Meta      interface{} `json:"meta,omitempty"`
	Errors    interface{} `json:"errors,omitempty"`
}

// WebResponses alias untuk kompatibilitas
type WebResponses = WebResponse
