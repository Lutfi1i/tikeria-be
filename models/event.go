package models

import (
	"time"

	"gorm.io/gorm"
)

type Event struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	OrganizerID uint           `gorm:"not null" json:"organizer_id"`
	Organizer   User           `gorm:"foreignKey:OrganizerID" json:"organizer,omitempty"`
	Title       string         `gorm:"type:varchar(150);not null" json:"title"`
	Description string         `gorm:"type:text" json:"description"`
	Location    string         `gorm:"type:varchar(150)" json:"location"`
	BannerURL   string         `gorm:"type:varchar(255)" json:"banner_url"`
	Category    string         `gorm:"type:varchar(50)" json:"category"`
	EventDate   time.Time      `json:"event_date"`
	Status      string         `gorm:"type:varchar(20);default:'published'" json:"status"`
	Tickets     []Ticket       `gorm:"foreignKey:EventID" json:"tickets,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

type Ticket struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	EventID   uint           `gorm:"not null" json:"event_id"`
	Name      string         `gorm:"type:varchar(100);not null" json:"name"`
	Price     float64        `gorm:"not null" json:"price"`
	Quota     int            `gorm:"not null" json:"quota"`
	Available int            `gorm:"not null" json:"available"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
