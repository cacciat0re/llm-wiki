package main

import (
	"fmt"
	"llm-wiki-server/internal/config"
	"llm-wiki-server/internal/repository"
	"llm-wiki-server/internal/route"
	"llm-wiki-server/internal/service"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	log.SetOutput(os.Stdout)

	// Load config
	cfg, err := config.Load("config.yml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := repository.Connect(cfg.DB)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// Repository
	wikiPageRepository := repository.NewWikiPageRepository(db)
	sourceRepository := repository.NewSourceRepository(db)

	// Service
	wikiPageService := service.NewWikiPageService(wikiPageRepository)
	sourceService := service.NewSourceService(sourceRepository)

	// Handler
	wikiPageHandler := route.NewWikiPageHandler(wikiPageService)
	sourceHandler := route.NewSourceHandler(sourceService)

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins: cfg.CORS.AllowedOrigins,
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type"},
		MaxAge:         300,
	}))

	wikiPageHandler.RegisterRoutes(router)
	sourceHandler.RegisterRoutes(router)

	log.Printf("Server started on port: %d", cfg.Server.Port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.Server.Port), router); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
