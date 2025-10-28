package models

import (
	"time"
)

type RefreshToken struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	UserID     uint      `json:"user_id" gorm:"index;not null"`
	TokenHash  string    `json:"-" gorm:"size:64;uniqueIndex;not null"`
	ParentID   *uint     `json:"parent_id,omitempty" gorm:"index"`
	ReplacedBy *uint     `json:"replaced_by,omitempty"`
	Revoked    bool      `json:"revoked" gorm:"default:false"`
	ExpiresAt  time.Time `json:"expires_at"`
	CreatedAt  time.Time `json:"created_at"`

	IPAddress *string `json:"ip_address,omitempty" gorm:"size:45"`
	UserAgent *string `json:"user_agent,omitempty" gorm:"size:512"`
}
