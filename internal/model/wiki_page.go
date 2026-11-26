package model

import "time"

type WikiPage struct {
	ID      int64       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Slug    string      `gorm:"column:slug;size:512;not null;uniqueIndex:uk_page_slug" json:"slug"`
	Title   string      `gorm:"column:title;size:512;not null" json:"title"`
	Alias   string      `gorm:"column:alias;type:text;not null;default:''" json:"alias"`
	Content string      `gorm:"column:content;type:text;not null;default:''" json:"content"`
	Tags    StringArray `gorm:"column:tags;type:text[]" json:"tags"`

	CreatedAt time.Time `gorm:"column:created_at;not null;autoCreateTime" json:"createdAt"`
	CreatedBy *string   `gorm:"column:created_by;type:uuid" json:"createdBy,omitempty"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null;autoUpdateTime" json:"updatedAt"`
	UpdatedBy *string   `gorm:"column:updated_by;type:uuid" json:"updatedBy,omitempty"`
}

type WikiPageSummary struct {
	ID        int64       `json:"id"`
	Slug      string      `json:"slug"`
	Title     string      `json:"title"`
	Alias     string      `gorm:"column:alias" json:"alias"`
	Tags      StringArray `gorm:"column:tags;type:text[]" json:"tags"`
	UpdatedAt time.Time   `json:"updatedAt"`
}

type WikiPageUpsertReq struct {
	Slug    string      `json:"slug"`
	Title   string      `json:"title"`
	Alias   string      `json:"alias"`
	Content string      `json:"content"`
	Tags    StringArray `json:"tags"`
}
