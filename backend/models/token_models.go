package models

import "time"

type RefreshToken struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	RefreshToken string    `gorm:"size:512;uniqueIndex;not null" json:"refresh_token"`
	ParentToken  string    `gorm:"size:512" json:"parent_token"`
	UserID       int       `gorm:"index;not null" json:"user_id"`
	Exp          int64     `gorm:"not null" json:"exp"`
	IPAddress    *string   `gorm:"size:45" json:"ip_address,omitempty"`
	UserAgent    *string   `gorm:"size:512" json:"user_agent,omitempty"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
}
