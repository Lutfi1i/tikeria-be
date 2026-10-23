package models

import (
	"time"

	"gorm.io/gorm"
)

type AuthSession struct {
	ID               string         `gorm:"primaryKey;type:varchar(64)" json:"id"`
	UserID           uint           `gorm:"not null;index" json:"user_id"`
	Role             string         `gorm:"type:varchar(20);not null" json:"role"`
	RefreshTokenHash string         `gorm:"type:varchar(255);not null" json:"-"`
	DeviceName       string         `gorm:"type:varchar(100)" json:"device_name"`
	UserAgent        string         `gorm:"type:text" json:"user_agent"`
	IPAddress        string         `gorm:"type:varchar(45)" json:"ip_address"`
	ExpiresAt        time.Time      `json:"expires_at"`
	RevokedAt        *time.Time     `json:"revoked_at"`
	LastUsedAt       time.Time      `json:"last_used_at"`
	IsActive         bool           `gorm:"default:true" json:"is_active"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

type ParticipantDetail struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	OrderID     uint           `gorm:"not null;uniqueIndex" json:"order_id"`
	FullNameKTP string         `gorm:"type:varchar(150);not null" json:"full_name_ktp"`
	Email       string         `gorm:"type:varchar(100);not null" json:"email"`
	Phone       string         `gorm:"type:varchar(30);not null" json:"phone"`
	NIK         string         `gorm:"type:varchar(20);not null" json:"-"`
	MaskedNIK   string         `gorm:"type:varchar(20)" json:"masked_nik"`
	BirthDate   time.Time      `json:"birth_date"`
	Gender      string         `gorm:"type:varchar(20)" json:"gender"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
