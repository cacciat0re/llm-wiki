package tool

import (
	"context"
	"errors"
	"fmt"
	"llm-wiki-server/internal/model"
	"llm-wiki-server/internal/service"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"gorm.io/gorm"
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
		Name:        "read_wiki_page",
		Description: "Read a wiki page by its unique slug. Returns the slug, title, and full Markdown content. Read a page before editing it so old_string matches the exact stored text.",
	}, h.ReadWikiPage)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "write_wiki_page",
		Description: "Create or fully overwrite a wiki page by slug. slug, title, and content are required. This replaces the whole page: omitting alias or tags clears them. Use edit_wiki_page for small, targeted changes.",
	}, h.WriteWikiPage)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "edit_wiki_page",
		Description: "Perform an exact string replacement in a wiki page's content, selected by slug. old_string must match exactly, including whitespace and indentation. Fails if old_string is not found or occurs multiple times without replace_all. Set replace_all to replace every occurrence. new_string may be empty to delete the matched text.",
	}, h.EditWikiPage)

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

type ReadInput struct {
	Slug string `json:"slug" jsonschema:"unique page slug, e.g. go-generics"`
}

type WriteInput struct {
	Slug    string   `json:"slug" jsonschema:"unique page slug, e.g. go-generics"`
	Title   string   `json:"title" jsonschema:"human-readable page title"`
	Alias   string   `json:"alias,omitempty" jsonschema:"optional comma-separated aliases; omitted clears existing aliases because write replaces the whole page"`
	Content string   `json:"content" jsonschema:"full Markdown page content"`
	Tags    []string `json:"tags,omitempty" jsonschema:"optional tags; omitted clears existing tags because write replaces the whole page"`
}

type EditInput struct {
	Slug       string `json:"slug" jsonschema:"slug of the page to edit"`
	OldString  string `json:"old_string" jsonschema:"exact text to replace; matched verbatim against content, including whitespace; must occur exactly once unless replace_all is true"`
	NewString  string `json:"new_string" jsonschema:"replacement text; may be empty to delete old_string"`
	ReplaceAll bool   `json:"replace_all,omitempty" jsonschema:"replace every occurrence instead of requiring a unique match"`
}

type ReadOutput struct {
	Slug    string `json:"slug"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

type WriteOutput struct {
	ID   int64  `json:"id"`
	Slug string `json:"slug"`
}

type EditOutput struct {
	Slug         string `json:"slug"`
	Replacements int    `json:"replacements"`
}

func (h *WikiPageHandler) ListWikiPages(ctx context.Context, _ *mcp.CallToolRequest, in ListInput) (*mcp.CallToolResult, service.PageResponse[model.WikiPageSummary], error) {
	keyword, page, size := in.Keyword, in.Page, in.Size

	resp, err := h.svc.List(ctx, keyword, page, size)
	if err != nil {
		return nil, service.PageResponse[model.WikiPageSummary]{}, err
	}
	return nil, resp, nil
}

func (h *WikiPageHandler) ReadWikiPage(ctx context.Context, _ *mcp.CallToolRequest, in ReadInput) (*mcp.CallToolResult, ReadOutput, error) {
	page, err := h.svc.GetBySlug(ctx, in.Slug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ReadOutput{}, fmt.Errorf("wiki page %q not found", in.Slug)
		}
		return nil, ReadOutput{}, err
	}

	return nil, ReadOutput{
		Slug:    page.Slug,
		Title:   page.Title,
		Content: page.Content,
	}, nil
}

func (h *WikiPageHandler) WriteWikiPage(ctx context.Context, _ *mcp.CallToolRequest, in WriteInput) (*mcp.CallToolResult, WriteOutput, error) {
	req := model.WikiPageUpsertReq{
		Slug:    in.Slug,
		Title:   in.Title,
		Alias:   in.Alias,
		Content: in.Content,
		Tags:    model.StringArray(in.Tags),
	}

	id, err := h.svc.UpsertBySlug(ctx, &req)
	if err != nil {
		return nil, WriteOutput{}, err
	}

	return nil, WriteOutput{ID: id, Slug: req.Slug}, nil
}

func (h *WikiPageHandler) EditWikiPage(ctx context.Context, _ *mcp.CallToolRequest, in EditInput) (*mcp.CallToolResult, EditOutput, error) {
	replacements, err := h.svc.Replace(ctx, in.Slug, in.OldString, in.NewString, in.ReplaceAll)
	if err != nil {
		return nil, EditOutput{}, err
	}

	return nil, EditOutput{Slug: in.Slug, Replacements: replacements}, nil
}

func (h *WikiPageHandler) DeleteWikiPage(ctx context.Context, _ *mcp.CallToolRequest, in GetInput) (*mcp.CallToolResult, struct{}, error) {
	if err := h.svc.Delete(ctx, in.ID); err != nil {
		return nil, struct{}{}, err
	}
	return nil, struct{}{}, nil
}
