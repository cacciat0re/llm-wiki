package main

import (
	"context"
	"flag"
	"log"
	"os"

	"llm-wiki-server/internal/config"
	"llm-wiki-server/internal/repository"
	"llm-wiki-server/internal/service"
	"llm-wiki-server/internal/tool"

	"github.com/jinzhu/configor"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	log.SetOutput(os.Stderr)

	configPath := flag.String("c", "config.yml", "path to config file")
	flag.Parse()

	if _, err := os.Stat(*configPath); err != nil {
		log.Fatalf("config file not found: %s", *configPath)
	}

	var cfg config.Configs
	// Silent: configor otherwise prints to stdout, which corrupts the MCP stdio stream.
	if err := configor.New(&configor.Config{Silent: true}).Load(&cfg, *configPath); err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

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

	t.RegisterTools(server)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
