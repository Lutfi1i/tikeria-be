package models

import (
	"time"

	"gorm.io/gorm"
)

type TicketHold struct {
	ID          uint               `gorm:"primaryKey" json:"id"`
	HoldCode    string             `gorm:"type:varchar(50);uniqueIndex;not null" json:"hold_code"`
	EventID     uint               `gorm:"not null;index" json:"event_id"`
	TicketID    uint               `gorm:"not null;index" json:"ticket_id"`
	UserID      uint               `gorm:"not null;index" json:"user_id"`
	Quantity    int                `gorm:"not null" json:"quantity"`
	Status      string             `gorm:"type:varchar(20);default:'Active';index" json:"status"` // Active, Confirmed, Released, Expired
	ExpiredAt   time.Time          `gorm:"not null;index" json:"expired_at"`
	ConfirmedAt *time.Time         `json:"confirmed_at,omitempty"`
	ReleasedAt  *time.Time         `json:"released_at,omitempty"`
	Participant *ParticipantDetail `gorm:"foreignKey:HoldID" json:"participant,omitempty"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
	DeletedAt   gorm.DeletedAt     `gorm:"index" json:"-"`
}
