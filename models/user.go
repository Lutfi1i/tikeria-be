package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	Name          string         `gorm:"type:varchar(100);not null" json:"name"`
	Email         string         `gorm:"type:varchar(100);uniqueIndex;not null" json:"email"`
	Password      string         `gorm:"type:varchar(255);not null" json:"-"`
	Role          string         `gorm:"type:varchar(20);default:'buyer'" json:"role"`
	WhatsApp      string         `gorm:"type:varchar(30)" json:"whatsapp,omitempty"`
	OrganizerType string         `gorm:"type:varchar(30);default:'school_campus'" json:"organizer_type,omitempty"`
	Institution   string         `gorm:"type:varchar(150)" json:"institution,omitempty"`
	City          string         `gorm:"type:varchar(100)" json:"city,omitempty"`
	Package       string         `gorm:"type:varchar(20);default:'basic'" json:"package,omitempty"`
	PackageSince  string         `gorm:"type:varchar(20)" json:"package_since,omitempty"`
	EmailVerified bool           `gorm:"default:false" json:"email_verified"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}
