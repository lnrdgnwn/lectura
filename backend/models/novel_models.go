package models

import (
	"time"
)

type Novel struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	AuthorID   uint      `gorm:"not null;index" json:"author_id"`
	Title      string    `gorm:"size:255;not null" json:"title"`
	Slug       string    `gorm:"size:255;uniqueIndex;not null" json:"slug"`
	Synopsis   *string   `gorm:"type:text" json:"synopsis,omitempty"`
	CoverImage *string   `gorm:"size:512" json:"cover_image,omitempty"`
	Status     string    `gorm:"size:20;default:'ongoing'" json:"status"`
	Genres     []Genre   `gorm:"many2many:novel_genres;constraint:OnDelete:CASCADE;" json:"genres,omitempty"`
	Tags       []Tag     `gorm:"many2many:novel_tags;constraint:OnDelete:CASCADE;" json:"tags,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
