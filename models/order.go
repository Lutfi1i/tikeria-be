package models

import (
	"time"

	"gorm.io/gorm"
)

type Order struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	OrderCode     string         `gorm:"type:varchar(50);uniqueIndex;not null" json:"order_code"`
	InvoiceNumber string         `gorm:"type:varchar(50);uniqueIndex" json:"invoice_number"`
	HoldID        uint           `gorm:"index" json:"hold_id"`
	UserID        uint           `gorm:"not null;index" json:"user_id"`
	User          User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	EventID       uint           `gorm:"not null;index" json:"event_id"`
	Event         Event          `gorm:"foreignKey:EventID" json:"event,omitempty"`
	TicketID      uint           `gorm:"not null" json:"ticket_id"`
	Ticket        Ticket         `gorm:"foreignKey:TicketID" json:"ticket,omitempty"`
	Quantity      int            `gorm:"not null" json:"quantity"`
	Subtotal      float64        `gorm:"type:decimal(12,2);default:0" json:"subtotal"`
	ServiceFee    float64        `gorm:"type:decimal(12,2);default:0" json:"service_fee"`
	TotalPrice    float64        `gorm:"type:decimal(12,2);not null" json:"total_price"` // keeps backward compatibility
	TotalAmount   float64        `gorm:"type:decimal(12,2);not null" json:"total_amount"`
	Status        string         `gorm:"type:varchar(30);default:'Pending Payment'" json:"status"` // 'Pending Payment', 'paid', 'cancelled', 'expired'
	ExpiredAt     *time.Time     `json:"expired_at,omitempty"`
	PaidAt        *time.Time     `json:"paid_at,omitempty"`
	Items         []OrderItem    `gorm:"foreignKey:OrderID" json:"items,omitempty"`
	Payment       *Payment       `gorm:"foreignKey:OrderID" json:"payment,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

type OrderItem struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	OrderID      uint           `gorm:"not null;index" json:"order_id"`
	TicketID     uint           `gorm:"not null;index" json:"ticket_id"`
	CategoryName string         `gorm:"type:varchar(100)" json:"category_name"`
	Price        float64        `gorm:"type:decimal(12,2);not null" json:"price"`
	Quantity     int            `gorm:"not null" json:"quantity"`
	Subtotal     float64        `gorm:"type:decimal(12,2);not null" json:"subtotal"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

type Payment struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	PaymentCode   string         `gorm:"type:varchar(50);uniqueIndex;not null" json:"payment_code"`
	OrderID       uint           `gorm:"not null;index" json:"order_id"`
	Status        string         `gorm:"type:varchar(30);default:'Waiting'" json:"status"` // Waiting, Success, Failed, Expired
	Amount        float64        `gorm:"type:decimal(12,2);not null" json:"amount"`
	PaymentMethod string         `gorm:"type:varchar(50);default:'Simulation'" json:"payment_method"`
	PaidAt        *time.Time     `json:"paid_at,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}
