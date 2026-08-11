package dto

import "time"

type CreateEventRequest struct {
	Title       string        `json:"title" binding:"required"`
	Description string        `json:"description" binding:"required"`
	Location    string        `json:"location" binding:"required"`
	BannerURL   string        `json:"banner_url"`
	Category    string        `json:"category" binding:"required"`
	EventDate   time.Time     `json:"event_date" binding:"required"`
	Tickets     []TicketInput `json:"tickets" binding:"required,dive"`
}

type TicketInput struct {
	Name  string  `json:"name" binding:"required"`
	Price float64 `json:"price" binding:"required,gte=0"`
	Quota int     `json:"quota" binding:"required,gt=0"`
}

type CreateOrderRequest struct {
	EventID  uint `json:"event_id" binding:"required"`
	TicketID uint `json:"ticket_id" binding:"required"`
	Quantity int  `json:"quantity" binding:"required,gt=0"`
}
