package repository

import (
	"context"
	"llm-wiki-server/internal/model"

	"gorm.io/gorm"
)

type IWikiPageRepository interface {
	List(ctx context.Context, index int, limit int) ([]model.WikiPageSummary, int64, error)
	Search(ctx context.Context, keyword string) ([]model.WikiPageSummary, error)
	Create(ctx context.Context, req *model.WikiPageUpsertReq) error
	GetById(ctx context.Context, id int64) (model.WikiPage, error)
	Update(ctx context.Context, id int64, req model.WikiPageUpsertReq) error
	Delete(ctx context.Context, id int64) error
}

type WikiPageRepository struct {
	db *gorm.DB
}

func NewWikiPageRepository(db *gorm.DB) IWikiPageRepository {
	return &WikiPageRepository{db: db}
}

func (r *WikiPageRepository) List(ctx context.Context, index int, limit int) ([]model.WikiPageSummary, int64, error) {
	db := r.db.WithContext(ctx).Model(&model.WikiPage{})

	// Get total number
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get by page
	var summaries []model.WikiPageSummary
	err := r.db.WithContext(ctx).
		Table("wiki_pages").
		Select("id", "slug", "title", "updated_at").
		Order("updated_at desc").
		Offset((index) * limit).
		Limit(limit).
		Scan(&summaries).Error
	if err != nil {
		return nil, 0, err
	}

	return summaries, total, nil
}

func (r *WikiPageRepository) Search(ctx context.Context, keyword string) ([]model.WikiPageSummary, error) {
	db := r.db.WithContext(ctx).Model(&model.WikiPage{})
	like := "%" + keyword + "%"
	db = db.Where("title ILIKE ? OR content ILIKE ?", like, like)

	var summaries []model.WikiPageSummary
	err := db.
		Select("id", "slug", "title", "updated_at").
		Order("updated_at desc").
		Scan(&summaries).Error
	if err != nil {
		return nil, err
	}
	return summaries, nil
}

func (r *WikiPageRepository) Create(ctx context.Context, req *model.WikiPageUpsertReq) error {
	page := model.WikiPage{
		Slug:    req.Slug,
		Title:   req.Title,
		Content: req.Content,
	}

	if err := r.db.WithContext(ctx).Create(&page).Error; err != nil {
		return err
	}

	return nil
}

func (r *WikiPageRepository) GetById(ctx context.Context, id int64) (model.WikiPage, error) {
	var page model.WikiPage
	if err := r.db.WithContext(ctx).
		Table("wiki_pages").
		Where("id = ?", id).
		Take(&page).Error; err != nil {
		return model.WikiPage{}, err
	}

	return page, nil
}

func (r *WikiPageRepository) Update(ctx context.Context, id int64, req model.WikiPageUpsertReq) error {
	result := r.db.WithContext(ctx).
		Model(&model.WikiPage{}).
		Where("id = ?", id).
		Updates(model.WikiPage{
			Title:   req.Title,
			Content: req.Content,
			Slug:    req.Slug,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *WikiPageRepository) Delete(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Delete(&model.WikiPage{}, id)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
