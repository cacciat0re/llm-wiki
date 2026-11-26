package repository

import (
	"context"
	"errors"
	"fmt"
	"llm-wiki-server/internal/model"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IWikiPageRepository interface {
	// List returns a page of summaries filtered by keyword. An empty keyword
	// matches every page. index is 0-based and limit is the page size.
	List(ctx context.Context, keyword string, index int, limit int) ([]model.WikiPageSummary, int64, error)
	Create(ctx context.Context, req *model.WikiPageUpsertReq) error
	GetById(ctx context.Context, id int64) (model.WikiPageDetail, error)
	GetBySlug(ctx context.Context, slug string) (model.WikiPageDetail, error)
	Update(ctx context.Context, id int64, req model.WikiPageUpsertReq) error
	UpsertBySlug(ctx context.Context, req *model.WikiPageUpsertReq) (int64, error)
	ReplaceContent(ctx context.Context, slug string, oldString string, newString string, replaceAll bool) (int, error)
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

	query.Select("id", "slug", "title", "alias", "tags").
		Order("updated_at desc").
		Offset(index * limit).
		Limit(limit)

	if err := query.Scan(&summaries).Error; err != nil {
		return nil, 0, err
	}

	return summaries, total, nil
}

func (r *WikiPageRepository) Create(ctx context.Context, req *model.WikiPageUpsertReq) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		page := model.WikiPage{
			Slug:    req.Slug,
			Title:   req.Title,
			Alias:   req.Alias,
			Content: req.Content,
			Tags:    req.Tags,
		}

		if err := tx.Create(&page).Error; err != nil {
			return err
		}

		return replacePageSources(tx, page.ID, req.SourceIDs)
	})
}

func (r *WikiPageRepository) GetById(ctx context.Context, id int64) (model.WikiPageDetail, error) {
	var page model.WikiPage
	if err := r.db.WithContext(ctx).
		Table("wiki_pages").
		Where("id = ?", id).
		Take(&page).Error; err != nil {
		return model.WikiPageDetail{}, err
	}

	return r.detail(ctx, page)
}

func (r *WikiPageRepository) GetBySlug(ctx context.Context, slug string) (model.WikiPageDetail, error) {
	var page model.WikiPage
	if err := r.db.WithContext(ctx).
		Table("wiki_pages").
		Where("slug = ?", slug).
		Take(&page).Error; err != nil {
		return model.WikiPageDetail{}, err
	}

	return r.detail(ctx, page)
}

func (r *WikiPageRepository) detail(ctx context.Context, page model.WikiPage) (model.WikiPageDetail, error) {
	sources := make([]model.Source, 0)
	if err := r.db.WithContext(ctx).
		Model(&model.Source{}).
		Joins("JOIN wiki_page_sources wps ON wps.source_id = sources.id").
		Where("wps.wiki_page_id = ?", page.ID).
		Order("sources.id desc").
		Find(&sources).Error; err != nil {
		return model.WikiPageDetail{}, err
	}

	return model.WikiPageDetail{
		ID:        page.ID,
		Slug:      page.Slug,
		Title:     page.Title,
		Alias:     page.Alias,
		Tags:      page.Tags,
		Content:   page.Content,
		UpdatedAt: page.UpdatedAt,
		Sources:   sources,
	}, nil
}

func (r *WikiPageRepository) Update(ctx context.Context, id int64, req model.WikiPageUpsertReq) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&model.WikiPage{}).
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

		return replacePageSources(tx, id, req.SourceIDs)
	})
}

func (r *WikiPageRepository) UpsertBySlug(ctx context.Context, req *model.WikiPageUpsertReq) (int64, error) {
	var id int64
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Raw(`
			INSERT INTO wiki_pages (slug, title, alias, content, tags)
			VALUES (?, ?, ?, ?, ?)
			ON CONFLICT (slug) DO UPDATE SET
				title = EXCLUDED.title, alias = EXCLUDED.alias,
				content = EXCLUDED.content, tags = EXCLUDED.tags,
				updated_at = NOW()
			RETURNING id`,
			req.Slug, req.Title, req.Alias, req.Content, req.Tags).Scan(&id).Error; err != nil {
			return err
		}

		// nil means the caller (e.g. MCP write without source_ids) did not
		// provide sources; keep the existing links. A non-nil slice replaces
		// them, and an empty slice clears them.
		if req.SourceIDs == nil {
			return nil
		}

		return replacePageSources(tx, id, req.SourceIDs)
	})
	return id, err
}

func replacePageSources(tx *gorm.DB, pageID int64, sourceIDs []int64) error {
	ids := make([]int64, 0, len(sourceIDs))
	seen := make(map[int64]struct{}, len(sourceIDs))
	for _, id := range sourceIDs {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}

	if len(ids) > 0 {
		var existing []int64
		if err := tx.Model(&model.Source{}).Where("id IN ?", ids).Pluck("id", &existing).Error; err != nil {
			return err
		}
		found := make(map[int64]struct{}, len(existing))
		for _, id := range existing {
			found[id] = struct{}{}
		}
		for _, id := range ids {
			if _, ok := found[id]; !ok {
				return fmt.Errorf("source %d not found: %w", id, gorm.ErrRecordNotFound)
			}
		}
	}

	if err := tx.Where("wiki_page_id = ?", pageID).Delete(&model.WikiPageSource{}).Error; err != nil {
		return err
	}
	if len(ids) == 0 {
		return nil
	}

	links := make([]model.WikiPageSource, 0, len(ids))
	for _, id := range ids {
		links = append(links, model.WikiPageSource{WikiPageID: pageID, SourceID: id})
	}
	return tx.Create(&links).Error
}

// ReplaceContent atomically replaces oldString in a page's content. It locks
// the row for the duration of the read-modify-write so concurrent edits cannot
// overwrite each other, and it touches only content and updated_at so title,
// alias, and tags are preserved.
func (r *WikiPageRepository) ReplaceContent(ctx context.Context, slug string, oldString string, newString string, replaceAll bool) (int, error) {
	replacements := 0

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var page model.WikiPage
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("slug = ?", slug).
			Take(&page).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("wiki page %q not found", slug)
			}
			return err
		}

		count := strings.Count(page.Content, oldString)
		if count == 0 {
			return errors.New("old_string not found in page content")
		}
		if count > 1 && !replaceAll {
			return fmt.Errorf("old_string occurs %d times in page content; pass replace_all=true or include more context", count)
		}

		content := strings.Replace(page.Content, oldString, newString, 1)
		replacements = 1
		if replaceAll {
			content = strings.ReplaceAll(page.Content, oldString, newString)
			replacements = count
		}

		if err := tx.Model(&model.WikiPage{}).
			Where("id = ?", page.ID).
			UpdateColumns(map[string]any{
				"content":    content,
				"updated_at": gorm.Expr("NOW()"),
			}).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return replacements, nil
}

func (r *WikiPageRepository) Backlinks(ctx context.Context, slug string) ([]model.WikiPageSummary, error) {
	var pages []model.WikiPageSummary
	err := r.db.WithContext(ctx).Model(&model.WikiPage{}).
		Where("content LIKE ?", "%[["+slug+"%").
		Select("id", "slug", "title", "alias", "tags").
		Where("slug <> ?", slug).
		Find(&pages).Error
	return pages, err
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
