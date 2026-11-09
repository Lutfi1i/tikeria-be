package dto

import "time"

type CreateHoldRequest struct {
	EventID          interface{} `json:"event_id" binding:"required"`
	TicketCategoryID interface{} `json:"ticket_category_id"`
	TicketID         interface{} `json:"ticket_id"`
	Quantity         int         `json:"quantity" binding:"required,min=1"`
}

type CreateHoldResponse struct {
	HoldID           string `json:"hold_id"`
	EventID          string `json:"event_id"`
	TicketCategoryID string `json:"ticket_category_id"`
	Quantity         int    `json:"quantity"`
	Status           string `json:"status"`
	ExpiredAt        string `json:"expired_at"`
	RemainingSeconds int    `json:"remaining_seconds"`
}

type SaveParticipantRequest struct {
	HoldID      interface{} `json:"hold_id" binding:"required"`
	FullName    string      `json:"full_name"`
	FullNameKTP string      `json:"full_name_ktp"`
	Email       string      `json:"email" binding:"required,email"`
	Phone       string      `json:"phone"`
	WhatsApp    string      `json:"whatsapp"`
	NIK         string      `json:"nik"`
	BirthDate   string      `json:"birth_date"`
	Gender      string      `json:"gender"`
	Institution string      `json:"institution"`
}

type ParticipantDetailData struct {
	FullName    string `json:"full_name,omitempty"`
	FullNameKTP string `json:"full_name_ktp,omitempty"`
	Email       string `json:"email"`
	Phone       string `json:"phone,omitempty"`
	WhatsApp    string `json:"whatsapp,omitempty"`
	MaskedNIK   string `json:"masked_nik,omitempty"`
	BirthDate   string `json:"birth_date,omitempty"`
	Gender      string `json:"gender,omitempty"`
	Institution string `json:"institution,omitempty"`
}

type SaveParticipantResponse struct {
	HoldID      string                `json:"hold_id"`
	Participant ParticipantDetailData `json:"participant"`
}

type CheckoutSummaryRequest struct {
	HoldID interface{} `json:"hold_id" binding:"required"`
}

type CheckoutEventSummary struct {
	EventID   string `json:"event_id"`
	Title     string `json:"title"`
	EventDate string `json:"event_date"`
	EventTime string `json:"event_time"`
	VenueName string `json:"venue_name"`
	City      string `json:"city"`
}

type CheckoutItemSummary struct {
	TicketCategoryID string  `json:"ticket_category_id"`
	CategoryName     string  `json:"category_name"`
	Price            float64 `json:"price"`
	Quantity         int     `json:"quantity"`
	Subtotal         float64 `json:"subtotal"`
}

type CheckoutSummaryResponse struct {
	HoldID      string               `json:"hold_id"`
	Event       CheckoutEventSummary `json:"event"`
	Item        CheckoutItemSummary  `json:"item"`
	ServiceFee  float64              `json:"service_fee"`
	TotalAmount float64              `json:"total_amount"`
	ExpiredAt   string               `json:"expired_at"`
}

type CreateOrderRequest struct {
	HoldID interface{} `json:"hold_id" binding:"required"`
}

type PaymentSummary struct {
	PaymentID     string  `json:"payment_id"`
	Status        string  `json:"status"`
	Amount        float64 `json:"amount"`
	PaymentMethod string  `json:"payment_method"`
}

type CreateOrderResponse struct {
	OrderID       string         `json:"order_id"`
	InvoiceNumber string         `json:"invoice_number"`
	Status        string         `json:"status"`
	Subtotal      float64        `json:"subtotal"`
	ServiceFee    float64        `json:"service_fee"`
	TotalAmount   float64        `json:"total_amount"`
	Payment       PaymentSummary `json:"payment"`
	ExpiredAt     string         `json:"expired_at"`
}

type OrderStatusResponse struct {
	OrderID       string  `json:"order_id"`
	InvoiceNumber string  `json:"invoice_number"`
	OrderStatus   string  `json:"order_status"`
	PaymentStatus string  `json:"payment_status"`
	TotalAmount   float64 `json:"total_amount"`
	ExpiredAt     string  `json:"expired_at"`
}

type SimulatePaymentRequest struct {
	PaymentMethod string `json:"payment_method"`
}

func FormatTimeRFC3339(t time.Time) string {
	return t.Format(time.RFC3339)
}
