package service

import (
	"context"
	"llm-wiki-server/internal/model"
	"llm-wiki-server/internal/repository"
)

type IWikiPageService interface {
	List(ctx context.Context, page int, size int) (PageResponse[model.WikiPageSummary], error)
	Search(ctx context.Context, keyword string) ([]model.WikiPageSummary, error)
	Create(ctx context.Context, req model.WikiPageUpsertReq) (model.WikiPage, error)
	GetById(ctx context.Context, id int64) (model.WikiPage, error)
	Update(ctx context.Context, id int64, req model.WikiPageUpsertReq) error
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

func (s *WikiPageService) List(ctx context.Context, page int, size int) (PageResponse[model.WikiPageSummary], error) {
	summaries, total, err := s.repo.List(ctx, page-1, size)
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

func (s *WikiPageService) Search(ctx context.Context, keyword string) ([]model.WikiPageSummary, error) {
	return s.repo.Search(ctx, keyword)
}

func (s *WikiPageService) Create(ctx context.Context, req model.WikiPageUpsertReq) (model.WikiPage, error) {
	return s.repo.Create(ctx, req)
}

func (s *WikiPageService) GetById(ctx context.Context, id int64) (model.WikiPage, error) {
	return s.repo.GetById(ctx, id)
}

func (s *WikiPageService) Update(ctx context.Context, id int64, req model.WikiPageUpsertReq) error {
	return s.repo.Update(ctx, id, req)
}
