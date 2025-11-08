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

type ChapterCreateRequest struct {
    NovelID uint   `json:"novel_id" example:"1"`
    Title   string `json:"title" example:"Chapter 1"`
    Content string `json:"content" example:"Isi chapter..."`
    Publish *bool  `json:"publish" example:"true"`
}

type ChapterUpdateRequest struct {
    Title   *string `json:"title" example:"Judul baru"`
    Content *string `json:"content" example:"Isi baru"`
    OrderNo *int    `json:"order_no" example:"2"`
    Publish *bool   `json:"publish" example:"true"`
}