package models

import (
	"time"
)

type User struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	Username       string         `gorm:"size:50;uniqueIndex;not null" json:"username"`
	Email          string         `gorm:"size:255;uniqueIndex;not null" json:"email"`
	PasswordHash   string         `gorm:"size:255;not null" json:"-"`
	Role           string         `gorm:"size:20;not null;default:'reader'" json:"role"`
	ProfilePicture *string        `json:"profile_picture,omitempty" gorm:"default:'http://127.0.0.1:3000/assets/IMG/profile_picture/Cukup.jpg'"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	RefreshToken   []RefreshToken `gorm:"constraint:OnDelete:CASCADE;" json:"-"`
}
