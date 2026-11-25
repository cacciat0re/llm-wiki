package repository

import (
	"context"
	"llm-wiki-server/internal/model"

	"gorm.io/gorm"
)

type IWikiPageRepository interface {
	// List returns a page of summaries filtered by keyword. An empty keyword
	// matches every page. index is 0-based and limit is the page size.
	List(ctx context.Context, keyword string, index int, limit int) ([]model.WikiPageSummary, int64, error)
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

func (r *WikiPageRepository) List(ctx context.Context, keyword string, index int, limit int) ([]model.WikiPageSummary, int64, error) {

	query := r.db.WithContext(ctx).Model(&model.WikiPageSummary{}).Table("wiki_pages")

	if keyword != "" {
		keyword = "%" + keyword + "%"
		query = query.Where("title ILIKE ? OR content ILIKE ? OR alias ILIKE ?", keyword, keyword, keyword)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var summaries []model.WikiPageSummary

	query.Select("id", "slug", "title", "alias", "tags", "updated_at").
		Order("updated_at desc").
		Offset(index * limit).
		Limit(limit)

	if err := query.Scan(&summaries).Error; err != nil {
		return nil, 0, err
	}

	return summaries, total, nil
}

func (r *WikiPageRepository) Create(ctx context.Context, req *model.WikiPageUpsertReq) error {
	page := model.WikiPage{
		Slug:    req.Slug,
		Title:   req.Title,
		Alias:   req.Alias,
		Content: req.Content,
		Tags:    req.Tags,
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

func (r *WikiPageRepository) GetBySlug(ctx context.Context, slug string) (model.WikiPage, error) {
	var page model.WikiPage
	if err := r.db.WithContext(ctx).
		Table("wiki_pages").
		Where("slug = ?", slug).
		Take(&page).Error; err != nil {
		return model.WikiPage{}, err
	}

	return page, nil
}

func (r *WikiPageRepository) Update(ctx context.Context, id int64, req model.WikiPageUpsertReq) error {
	if req.Tags == nil {
		req.Tags = model.StringArray{}
	}

	result := r.db.WithContext(ctx).
		Model(&model.WikiPage{}).
		Where("id = ?", id).
		// Select forces zero values to be written, so an empty alias/tags
		// actually clears the stored value instead of being skipped.
		Select("slug", "title", "alias", "content", "tags").
		Updates(model.WikiPage{
			Title:   req.Title,
			Alias:   req.Alias,
			Content: req.Content,
			Slug:    req.Slug,
			Tags:    req.Tags,
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
