package service

import (
	"context"
	"errors"
	"fmt"
	"llm-wiki-server/internal/model"
	"llm-wiki-server/internal/repository"
	"strings"

	"gorm.io/gorm"
)

type ISourceService interface {
	List(ctx context.Context, keyword string, page int, size int) (PageResponse[model.Source], error)
	Get(ctx context.Context, id int64) (model.Source, error)
	UpsertByOrigin(ctx context.Context, origin string, title string, summary string) (model.Source, error)
	Delete(ctx context.Context, id int64) error
	ListPagesBySource(ctx context.Context, sourceId int64) ([]model.WikiPageSummary, error)
}

type SourceService struct {
	repo repository.ISourceRepository
}

func NewSourceService(repo repository.ISourceRepository) ISourceService {
	return &SourceService{repo: repo}
}

func (s *SourceService) List(ctx context.Context, keyword string, page int, size int) (PageResponse[model.Source], error) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	if size > 100 {
		size = 100
	}

	sources, total, err := s.repo.List(ctx, keyword, page-1, size)
	if err != nil {
		return PageResponse[model.Source]{}, err
	}
	totalPages := int((total + int64(size) - 1) / int64(size))

	return PageResponse[model.Source]{
		Content:    sources,
		Page:       page,
		Size:       size,
		Total:      total,
		TotalPages: totalPages,
		IsLast:     page >= totalPages,
	}, nil
}

func (s *SourceService) Get(ctx context.Context, id int64) (model.Source, error) {
	return s.repo.GetById(ctx, id)
}

func (s *SourceService) UpsertByOrigin(ctx context.Context, origin string, title string, summary string) (model.Source, error) {
	origin = strings.TrimSpace(origin)
	if origin == "" {
		return model.Source{}, errors.New("origin is required")
	}

	return s.repo.UpsertByOrigin(ctx, origin, title, summary)
}

func (s *SourceService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func (s *SourceService) ListPagesBySource(ctx context.Context, sourceId int64) ([]model.WikiPageSummary, error) {
	if _, err := s.sourceById(ctx, sourceId); err != nil {
		return nil, err
	}

	return s.repo.ListPagesBySourceId(ctx, sourceId)
}

func (s *SourceService) sourceById(ctx context.Context, id int64) (model.Source, error) {
	source, err := s.repo.GetById(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.Source{}, fmt.Errorf("source %d not found: %w", id, err)
		}
		return model.Source{}, err
	}

	return source, nil
}
