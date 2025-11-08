package models

import "time"

type User struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	Username       string         `gorm:"size:50;uniqueIndex;not null" json:"username"`
	Email          string         `gorm:"size:255;uniqueIndex;not null" json:"email"`
	PasswordHash   string         `gorm:"size:255;not null" json:"-"`
	Role           string         `gorm:"size:20;not null;default:'user'" json:"role"`
	ProfilePicture *string        `gorm:"default:'http://127.0.0.1:3000/assets/IMG/profile_picture/default_picture.jpg'" json:"profile_picture,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	RefreshToken   []RefreshToken `gorm:"constraint:OnDelete:CASCADE;" json:"-"`
}
type RegisterRequest struct {
	Username string `json:"username" example:"johndoe"`
	Email    string `json:"email" example:"john@example.com"`
	Password string `json:"password" example:"secret123"`
}
type LoginRequest struct {
	Email    string `json:"email" example:"john@example.com"`
	Password string `json:"password" example:"secret123"`
}
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" example:"<encrypted_refresh_token>"`
}