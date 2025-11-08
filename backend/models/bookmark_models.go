package models

import "time"

type Bookmark struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index:idx_bookmarks_user_novel;not null" json:"user_id"`
	NovelID   uint      `gorm:"index:idx_bookmarks_user_novel;not null" json:"novel_id"`
	CreatedAt time.Time `json:"created_at"`

	Novel Novel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:NovelID;references:ID" json:"-"`
}
