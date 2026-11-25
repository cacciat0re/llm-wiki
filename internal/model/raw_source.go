package model

import "time"

// RawSource reference content of wiki page
type RawSource struct {
	ID int64 `json:"id"`
	// Origin local file path or URL
	Origin string `json:"origin"`
	Slug   string `json:"slug"`
	Title  string `json:"title"`

	CreatedAt time.Time `gorm:"column:created_at;not null;autoCreateTime" json:"createdAt"`
}
