package models

import (
	"time"
)

type Genre struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Name       string    `gorm:"size:100;uniqueIndex;not null" json:"name"`
	Slug       string    `gorm:"size:120;uniqueIndex;not null" json:"slug"`
	ShowOnHome bool      `json:"show_on_home" gorm:"default:false"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
