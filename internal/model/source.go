package model

// Source is a global raw source shared by multiple wiki pages.
type Source struct {
	ID int64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	// Origin local file path or URL. It is the identity key and is not checked
	// for existence: the file may live on another machine or may have moved.
	Origin  string `gorm:"column:origin;size:2048;not null;uniqueIndex:uk_sources_origin" json:"origin"`
	Title   string `gorm:"column:title;size:512;not null;default:''" json:"title"`
	Summary string `gorm:"column:summary;type:text;not null;default:''" json:"summary"`
}
