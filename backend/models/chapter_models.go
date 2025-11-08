package models

import (
	"time"
)

type Chapter struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	NovelID     uint       `gorm:"not null;index" json:"novel_id"`
	Novel       Novel      `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	OrderNo     int        `gorm:"not null" json:"order_no"`
	Title       *string    `gorm:"size:255" json:"title,omitempty"`
	Content     *string    `gorm:"type:mediumtext" json:"content,omitempty"`
	PublishedAt *time.Time `gorm:"index" json:"published_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}