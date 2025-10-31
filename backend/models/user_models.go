package models

import (
	"time"
)

type User struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	Username       string         `gorm:"size:50;uniqueIndex;not null" json:"username"`
	Email          string         `gorm:"size:255;uniqueIndex;not null" json:"email"`
	PasswordHash   string         `gorm:"size:255;not null" json:"-"`
	Role           string         `gorm:"size:20;not null;default:'READER'" json:"role"`
	ProfilePicture *string        `gorm:"size:512" json:"profile_picture,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	RefreshTokens  []RefreshToken `gorm:"constraint:OnDelete:CASCADE;" json:"-"`
}
