package models

import (
	"time"

	"gorm.io/gorm"
)

type Event struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	OrganizerID   uint           `gorm:"not null" json:"organizer_id"`
	Organizer     User           `gorm:"foreignKey:OrganizerID" json:"organizer,omitempty"`
	Title         string         `gorm:"type:varchar(150);not null" json:"title"`
	Description   string         `gorm:"type:text" json:"description"`
	Location      string         `gorm:"type:varchar(150)" json:"location"`
	City          string         `gorm:"type:varchar(100)" json:"city,omitempty"`
	VenueName     string         `gorm:"type:varchar(150)" json:"venue_name,omitempty"`
	VenueAddress  string         `gorm:"type:text" json:"venue_address,omitempty"`
	BannerURL     string         `gorm:"type:varchar(255)" json:"banner_url"`
	Category      string         `gorm:"type:varchar(50)" json:"category"`
	EventTemplate string         `gorm:"type:varchar(30);default:'pensi'" json:"event_template"`
	PICName       string         `gorm:"type:varchar(100)" json:"pic_name,omitempty"`
	PICWhatsApp   string         `gorm:"type:varchar(30)" json:"pic_whatsapp,omitempty"`
	EventDate     time.Time      `json:"event_date"`
	Status        string         `gorm:"type:varchar(30);default:'Draft'" json:"status"`
	RejectReason  string         `gorm:"type:text" json:"reject_reason,omitempty"`
	ReviewedBy    string         `gorm:"type:varchar(100)" json:"reviewed_by,omitempty"`
	ReviewedAt    *time.Time     `json:"reviewed_at,omitempty"`
	PublishedAt   *time.Time     `json:"published_at,omitempty"`
	Tickets       []Ticket       `gorm:"foreignKey:EventID" json:"tickets,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

type Ticket struct {
	ID                      uint           `gorm:"primaryKey" json:"id"`
	EventID                 uint           `gorm:"not null" json:"event_id"`
	Name                    string         `gorm:"type:varchar(100);not null" json:"name"`
	CategoryName            string         `gorm:"type:varchar(100)" json:"category_name,omitempty"`
	Description             string         `gorm:"type:text" json:"description,omitempty"`
	Price                   float64        `gorm:"not null" json:"price"`
	Quota                   int            `gorm:"not null" json:"quota"`
	SoldQuantity            int            `gorm:"default:0" json:"sold_quantity"`
	Available               int            `gorm:"not null" json:"available"`
	MaxTicketPerTransaction int            `gorm:"default:5" json:"max_ticket_per_transaction"`
	IsActive                bool           `gorm:"default:true" json:"is_active"`
	CreatedAt               time.Time      `json:"created_at"`
	UpdatedAt               time.Time      `json:"updated_at"`
	DeletedAt               gorm.DeletedAt `gorm:"index" json:"-"`
}
