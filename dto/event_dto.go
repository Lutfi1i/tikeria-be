package dto

import "time"

type CreateEventRequest struct {
	Title         string        `json:"title" binding:"required"`
	Description   string        `json:"description" binding:"required"`
	Location      string        `json:"location"`
	City          string        `json:"city"`
	VenueName     string        `json:"venue_name"`
	VenueAddress  string        `json:"venue_address"`
	BannerURL     string        `json:"banner_url"`
	Category      string        `json:"category"`
	EventTemplate string        `json:"event_template"`
	PICName       string        `json:"pic_name"`
	PICWhatsApp   string        `json:"pic_whatsapp"`
	EventDate     time.Time     `json:"event_date" binding:"required"`
	Status        string        `json:"status"`
	Tickets       []TicketInput `json:"tickets" binding:"omitempty,dive"`
}

type UpdateEventRequest struct {
	Title         string    `json:"title" binding:"required"`
	Description   string    `json:"description" binding:"required"`
	Location      string    `json:"location"`
	City          string    `json:"city"`
	VenueName     string    `json:"venue_name"`
	VenueAddress  string    `json:"venue_address"`
	BannerURL     string    `json:"banner_url"`
	Category      string    `json:"category"`
	EventTemplate string    `json:"event_template"`
	PICName       string    `json:"pic_name"`
	PICWhatsApp   string    `json:"pic_whatsapp"`
	EventDate     time.Time `json:"event_date"`
}

type TicketInput struct {
	Name                    string  `json:"name"`
	CategoryName            string  `json:"category_name"`
	Description             string  `json:"description"`
	Price                   float64 `json:"price" binding:"gte=0"`
	Quota                   int     `json:"quota" binding:"required,gt=0"`
	MaxTicketPerTransaction int     `json:"max_ticket_per_transaction"`
	IsActive                *bool   `json:"is_active"`
}

type RejectEventRequest struct {
	RejectReason string `json:"reject_reason" binding:"required"`
}

type CreateOrderRequest struct {
	EventID  uint `json:"event_id" binding:"required"`
	TicketID uint `json:"ticket_id" binding:"required"`
	Quantity int  `json:"quantity" binding:"required,gt=0"`
}
