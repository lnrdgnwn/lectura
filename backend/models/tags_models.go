package models

import "time"

type Tag struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:100;not null" json:"name"`
	Slug      string    `gorm:"size:120;uniqueIndex;not null" json:"slug"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type NovelTag struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	NovelID   uint      `gorm:"index;not null" json:"novel_id"`
	TagID     uint      `gorm:"index;not null" json:"tag_id"`
	CreatedAt time.Time `json:"created_at"`
	// optional preload:
	Novel *Novel `gorm:"foreignKey:NovelID" json:"novel,omitempty"`
	Tag   *Tag   `gorm:"foreignKey:TagID" json:"tag,omitempty"`
}
