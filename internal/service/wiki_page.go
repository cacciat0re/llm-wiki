package service

import (
	"context"
	"errors"
	"llm-wiki-server/internal/model"
	"llm-wiki-server/internal/repository"
)

type IWikiPageService interface {
	List(ctx context.Context, keyword string, page int, size int) (PageResponse[model.WikiPageSummary], error)
	Create(ctx context.Context, req *model.WikiPageUpsertReq) error
	GetById(ctx context.Context, id int64) (model.WikiPage, error)
	GetBySlug(ctx context.Context, slug string) (model.WikiPage, error)
	Update(ctx context.Context, id int64, req model.WikiPageUpsertReq) error
	UpsertBySlug(ctx context.Context, req *model.WikiPageUpsertReq) (int64, error)
	Replace(ctx context.Context, slug string, oldString string, newString string, replaceAll bool) (int, error)
	Delete(ctx context.Context, id int64) error
}

type PageResponse[T any] struct {
	Content    []T   `json:"content"`
	Page       int   `json:"page"`
	Size       int   `json:"size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"totalPages"`
	IsLast     bool  `json:"isLast"`
}

type WikiPageService struct {
	repo repository.IWikiPageRepository
}

func NewWikiPageService(repo repository.IWikiPageRepository) IWikiPageService {
	return &WikiPageService{repo: repo}
}

func (s *WikiPageService) List(ctx context.Context, keyword string, page int, size int) (PageResponse[model.WikiPageSummary], error) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	if size > 100 {
		size = 100
	}

	summaries, total, err := s.repo.List(ctx, keyword, page-1, size)
	if err != nil {
		return PageResponse[model.WikiPageSummary]{}, err
	}
	totalPages := int((total + int64(size) - 1) / int64(size))

	return PageResponse[model.WikiPageSummary]{
		Content:    summaries,
		Page:       page,
		Size:       size,
		Total:      total,
		TotalPages: totalPages,
		IsLast:     page >= totalPages,
	}, nil
}

func (s *WikiPageService) Create(ctx context.Context, req *model.WikiPageUpsertReq) error {
	return s.repo.Create(ctx, req)
}

func (s *WikiPageService) GetById(ctx context.Context, id int64) (model.WikiPage, error) {
	return s.repo.GetById(ctx, id)
}

func (s *WikiPageService) GetBySlug(ctx context.Context, slug string) (model.WikiPage, error) {
	return s.repo.GetBySlug(ctx, slug)
}

//func (s *WikiPageService) GetBackLinks(ctx context.Context, id int64) ([]model.WikiPageSummary, error) {
//	return s.repo.
//}

func (s *WikiPageService) Update(ctx context.Context, id int64, req model.WikiPageUpsertReq) error {
	return s.repo.Update(ctx, id, req)
}

func (s *WikiPageService) UpsertBySlug(ctx context.Context, req *model.WikiPageUpsertReq) (int64, error) {
	if req.Slug == "" {
		return 0, errors.New("slug is required")
	}
	return s.repo.UpsertBySlug(ctx, req)
}

func (s *WikiPageService) Replace(ctx context.Context, slug string, oldString string, newString string, replaceAll bool) (int, error) {
	if slug == "" {
		return 0, errors.New("slug is required")
	}
	if oldString == "" {
		return 0, errors.New("old_string must not be empty")
	}
	return s.repo.ReplaceContent(ctx, slug, oldString, newString, replaceAll)
}

func (s *WikiPageService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
