package tool

import (
	"context"
	"llm-wiki-server/internal/model"
	"llm-wiki-server/internal/service"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type WikiPageHandler struct {
	svc service.IWikiPageService
}

func NewWikiPageHandler(svc service.IWikiPageService) *WikiPageHandler {
	return &WikiPageHandler{svc: svc}
}

func (h *WikiPageHandler) RegisterTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_wiki_pages",
		Description: "List or search wiki pages, most recently updated first. Provide keyword to match titles and content; omit it to list all. Results contain page metadata only, not content. Pagination is 1-based.",
	}, h.ListWikiPages)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_wiki_page",
		Description: "Create a new wiki page. slug, title, and content are required; tags are optional. Slugs must be unique.",
	}, h.CreateWikiPage)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_wiki_page",
		Description: "Replace an existing wiki page's slug, title, content, and tags, selected by numeric id. This is a full replacement, not a partial patch.",
	}, h.UpdateWikiPage)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_wiki_page",
		Description: "Permanently delete a wiki page by numeric id. The page cannot be recovered.",
	}, h.DeleteWikiPage)
}

type ListInput struct {
	Keyword string `json:"keyword,omitempty" jsonschema:"optional keyword to match titles and content; omit to list all pages"`
	Page    int    `json:"page,omitempty" jsonschema:"1-based page number"`
	Size    int    `json:"size,omitempty" jsonschema:"page size, should below 100"`
}

type GetInput struct {
	ID int64 `json:"id" jsonschema:"page id"`
}

type CreateInput struct {
	Slug    string   `json:"slug" jsonschema:"unique page slug, e.g. go-generics"`
	Title   string   `json:"title" jsonschema:"human-readable page title"`
	Alias   string   `json:"alias,omitempty" jsonschema:"optional comma-separated aliases (e.g. 'c++, cpp'); used only for fuzzy matching"`
	Content string   `json:"content" jsonschema:"full Markdown page content"`
	Tags    []string `json:"tags,omitempty" jsonschema:"optional tags; on create omit for none, on update omit to keep existing tags"`
}

type UpdateInput struct {
	ID int64 `json:"id" jsonschema:"page id"`
	CreateInput
}

func (h *WikiPageHandler) ListWikiPages(ctx context.Context, _ *mcp.CallToolRequest, in ListInput) (*mcp.CallToolResult, service.PageResponse[model.WikiPageSummary], error) {
	keyword, page, size := in.Keyword, in.Page, in.Size

	resp, err := h.svc.List(ctx, keyword, page, size)
	if err != nil {
		return nil, service.PageResponse[model.WikiPageSummary]{}, err
	}
	return nil, resp, nil
}

func (h *WikiPageHandler) CreateWikiPage(ctx context.Context, _ *mcp.CallToolRequest, in CreateInput) (*mcp.CallToolResult, struct{}, error) {
	req := model.WikiPageUpsertReq{
		Slug:    in.Slug,
		Title:   in.Title,
		Alias:   in.Alias,
		Content: in.Content,
		Tags:    model.StringArray(in.Tags),
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
		Alias:   in.Alias,
		Content: in.Content,
		Tags:    model.StringArray(in.Tags),
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
