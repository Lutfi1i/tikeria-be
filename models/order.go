package models

import (
	"time"

	"gorm.io/gorm"
)

type Order struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	OrderCode  string         `gorm:"type:varchar(50);uniqueIndex;not null" json:"order_code"`
	UserID     uint           `gorm:"not null" json:"user_id"`
	User       User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	EventID    uint           `gorm:"not null" json:"event_id"`
	Event      Event          `gorm:"foreignKey:EventID" json:"event,omitempty"`
	TicketID   uint           `gorm:"not null" json:"ticket_id"`
	Ticket     Ticket         `gorm:"foreignKey:TicketID" json:"ticket,omitempty"`
	Quantity   int            `gorm:"not null" json:"quantity"`
	TotalPrice float64        `gorm:"not null" json:"total_price"`
	Status     string         `gorm:"type:varchar(20);default:'paid'" json:"status"` // 'pending', 'paid', 'cancelled'
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}