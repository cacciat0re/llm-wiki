package model

// WikiPageSource links a wiki page to a global source.
type WikiPageSource struct {
	WikiPageID int64 `gorm:"column:wiki_page_id;primaryKey" json:"wikiPageId"`
	SourceID   int64 `gorm:"column:source_id;primaryKey" json:"sourceId"`
}
