package main

import (
	"context"
	"flag"
	"llm-wiki-server/internal/config"
	"llm-wiki-server/internal/repository"
	"llm-wiki-server/internal/service"
	"llm-wiki-server/internal/tool"
	"log"
	"os"
	"path/filepath"

	"github.com/jinzhu/configor"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	log.SetOutput(os.Stderr)

	cwd, _ := os.Getwd()
	log.Printf("[llm-wiki-mcp] starting: args=%q cwd=%s", os.Args, cwd)

	configPath := flag.String("config", "", "path to config file (default: $LLM_WIKI_CONFIG or config.yml next to the binary's parent dir)")
	flag.Parse()

	path := *configPath
	if path == "" {
		path = os.Getenv("LLM_WIKI_CONFIG")
	}
	if path == "" {
		if exe, err := os.Executable(); err == nil {
			path = filepath.Join(filepath.Dir(filepath.Dir(exe)), "config.yml")
		}
	}
	if _, err := os.Stat(path); err != nil {
		log.Fatalf("[llm-wiki-mcp] config file not found: %q (args=%q cwd=%s)", path, os.Args, cwd)
	}
	log.Printf("[llm-wiki-mcp] using config: %s", path)

	var cfg config.Configs
	if err := configor.New(&configor.Config{Silent: true}).Load(&cfg, path); err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	log.Printf("[llm-wiki-mcp] db target: %s@%s:%d/%s", cfg.DB.Username, cfg.DB.Host, cfg.DB.Port, cfg.DB.DatabaseName)

	db, err := repository.Connect(cfg.DB)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	wikiPageService := service.NewWikiPageService(repository.NewWikiPageRepository(db))
	t := tool.NewWikiPageHandler(wikiPageService)

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "llm-wiki-mcp",
		Version: "0.0.1",
	}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_wiki_pages",
		Description: "List wiki pages, most recently updated first. Results contain page metadata only, not content. Pagination is 1-based.",
	}, t.ListWikiPages)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "search_wiki_pages",
		Description: "Search page titles and content for a keyword. Returns matching page metadata, most recently updated first.",
	}, t.SearchWikiPage)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_wiki_page",
		Description: "Create a new wiki page. slug, title, and content are required; tags are optional. Slugs must be unique.",
	}, t.CreateWikiPage)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_wiki_page",
		Description: "Replace an existing wiki page's slug, title, content, and tags, selected by numeric id. This is a full replacement, not a partial patch.",
	}, t.UpdateWikiPage)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_wiki_page",
		Description: "Permanently delete a wiki page by numeric id. The page cannot be recovered.",
	}, t.DeleteWikiPage)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
