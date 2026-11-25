package tool

import (
	"context"
	"llm-wiki-server/internal/model"
	"llm-wiki-server/internal/service"

	"github.com/lib/pq"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type WikiPageHandler struct {
	svc service.IWikiPageService
}

func NewWikiPageHandler(svc service.IWikiPageService) *WikiPageHandler {
	return &WikiPageHandler{svc: svc}
}

type ListInput struct {
	Page int `json:"page,omitempty" jsonschema:"1-based page number"`
	Size int `json:"size,omitempty" jsonschema:"page size, should below 100"`
}

type SearchInput struct {
	Keyword string `json:"keyword" jsonschema:"keyword to search in titles and content"`
}

// SearchOutput wraps results in an object because MCP requires tool
// outputSchema to be an object schema (a bare array is rejected by strict clients).
type SearchOutput struct {
	Items []model.WikiPageSummary `json:"items" jsonschema:"matching page summaries, most recently updated first"`
}

type GetInput struct {
	ID int64 `json:"id" jsonschema:"page id"`
}

type CreateInput struct {
	Slug    string   `json:"slug" jsonschema:"unique page slug, e.g. go-generics"`
	Title   string   `json:"title" jsonschema:"human-readable page title"`
	Content string   `json:"content" jsonschema:"full Markdown page content"`
	Tags    []string `json:"tags,omitempty" jsonschema:"optional tags; on create omit for none, on update omit to keep existing tags"`
}

type UpdateInput struct {
	ID int64 `json:"id" jsonschema:"page id"`
	CreateInput
}

func (h *WikiPageHandler) ListWikiPages(ctx context.Context, _ *mcp.CallToolRequest, in ListInput) (*mcp.CallToolResult, service.PageResponse[model.WikiPageSummary], error) {
	page, size := in.Page, in.Size

	resp, err := h.svc.List(ctx, page, size)
	if err != nil {
		return nil, service.PageResponse[model.WikiPageSummary]{}, err
	}
	return nil, resp, nil
}

func (h *WikiPageHandler) SearchWikiPage(ctx context.Context, _ *mcp.CallToolRequest, in SearchInput) (*mcp.CallToolResult, SearchOutput, error) {
	resp, err := h.svc.Search(ctx, in.Keyword)
	if err != nil {
		return nil, SearchOutput{}, err
	}
	return nil, SearchOutput{Items: resp}, nil
}

func (h *WikiPageHandler) CreateWikiPage(ctx context.Context, _ *mcp.CallToolRequest, in CreateInput) (*mcp.CallToolResult, struct{}, error) {
	req := model.WikiPageUpsertReq{
		Slug:    in.Slug,
		Title:   in.Title,
		Content: in.Content,
		Tags:    pq.StringArray(in.Tags),
	}
	if err := h.svc.Create(ctx, &req); err != nil {
		return nil, struct{}{}, err
	}
	return nil, struct{}{}, nil
}

func (h *WikiPageHandler) UpdateWikiPage(ctx context.Context, _ *mcp.CallToolRequest, in UpdateInput) (*mcp.CallToolResult, struct{}, error) {
	req := model.WikiPageUpsertReq{
		Slug:    in.Slug,
		Title:   in.Title,
		Content: in.Content,
		Tags:    pq.StringArray(in.Tags),
	}
	if err := h.svc.Update(ctx, in.ID, req); err != nil {
		return nil, struct{}{}, err
	}
	return nil, struct{}{}, nil
}

func (h *WikiPageHandler) DeleteWikiPage(ctx context.Context, _ *mcp.CallToolRequest, in GetInput) (*mcp.CallToolResult, struct{}, error) {
	if err := h.svc.Delete(ctx, in.ID); err != nil {
		return nil, struct{}{}, err
	}
	return nil, struct{}{}, nil
}
