package repository

import (
	"context"
	"llm-wiki-server/internal/model"

	"gorm.io/gorm"
)

type ISourceRepository interface {
	List(ctx context.Context, keyword string, index int, limit int) ([]model.Source, int64, error)
	GetById(ctx context.Context, id int64) (model.Source, error)
	UpsertByOrigin(ctx context.Context, origin string, title string, summary string) (model.Source, error)
	Delete(ctx context.Context, id int64) error
	ListPagesBySourceId(ctx context.Context, sourceId int64) ([]model.WikiPageSummary, error)
}

type SourceRepository struct {
	db *gorm.DB
}

func NewSourceRepository(db *gorm.DB) ISourceRepository {
	return &SourceRepository{db: db}
}

func (r *SourceRepository) List(ctx context.Context, keyword string, index int, limit int) ([]model.Source, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.Source{})

	if keyword != "" {
		keyword = "%" + keyword + "%"
		query = query.Where("origin ILIKE ? OR title ILIKE ? OR summary ILIKE ?", keyword, keyword, keyword)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	sources := make([]model.Source, 0)
	if err := query.Order("id desc").Offset(index * limit).Limit(limit).Find(&sources).Error; err != nil {
		return nil, 0, err
	}

	return sources, total, nil
}

func (r *SourceRepository) GetById(ctx context.Context, id int64) (model.Source, error) {
	var source model.Source
	if err := r.db.WithContext(ctx).
		Table("sources").
		Where("id = ?", id).
		Take(&source).Error; err != nil {
		return model.Source{}, err
	}

	return source, nil
}

func (r *SourceRepository) UpsertByOrigin(ctx context.Context, origin string, title string, summary string) (model.Source, error) {
	var source model.Source
	err := r.db.WithContext(ctx).Raw(`
		INSERT INTO sources (origin, title, summary)
		VALUES (?, ?, ?)
		ON CONFLICT (origin) DO UPDATE SET
			title = EXCLUDED.title, summary = EXCLUDED.summary
		RETURNING id, origin, title, summary`,
		origin, title, summary).Scan(&source).Error
	return source, err
}

func (r *SourceRepository) Delete(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Delete(&model.Source{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *SourceRepository) ListPagesBySourceId(ctx context.Context, sourceId int64) ([]model.WikiPageSummary, error) {
	pages := make([]model.WikiPageSummary, 0)
	err := r.db.WithContext(ctx).
		Model(&model.WikiPageSummary{}).
		Table("wiki_pages p").
		Select("p.id", "p.slug", "p.title", "p.alias", "p.tags").
		Joins("JOIN wiki_page_sources wps ON wps.wiki_page_id = p.id").
		Where("wps.source_id = ?", sourceId).
		Order("p.updated_at desc, p.id desc").
		Scan(&pages).Error
	return pages, err
}
