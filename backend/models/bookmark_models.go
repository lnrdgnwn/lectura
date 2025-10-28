package models

import (
	"time"
)

type Bookmark struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	NovelID   uint      `gorm:"index;not null" json:"novel_id"`
	CreatedAt time.Time `json:"created_at"`
}
