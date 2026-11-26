package route

import (
	"encoding/json"
	"errors"
	"llm-wiki-server/internal/model"
	"llm-wiki-server/internal/service"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type WikiPageHandler struct {
	svc service.IWikiPageService
}

func NewWikiPageHandler(svc service.IWikiPageService) *WikiPageHandler {
	return &WikiPageHandler{
		svc: svc,
	}
}

func (h *WikiPageHandler) RegisterRoutes(router chi.Router) {
	router.Route("/api/wiki", func(r chi.Router) {
		r.Get("/", h.List)
		r.Get("/{slug}", h.GetBySlug)
		r.Post("/", h.Create)
		r.Put("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
	})
}

func (h *WikiPageHandler) List(w http.ResponseWriter, r *http.Request) {
	keyword := r.URL.Query().Get("keyword")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	size, _ := strconv.Atoi(r.URL.Query().Get("size"))

	resp, err := h.svc.List(r.Context(), keyword, page, size)
	if err != nil {
		log.Printf("failed to list wiki pages: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("failed to write response: %v", err)
	}
}

func (h *WikiPageHandler) Create(w http.ResponseWriter, r *http.Request) {
	var page model.WikiPageUpsertReq
	if err := json.NewDecoder(r.Body).Decode(&page); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.svc.Create(r.Context(), &page); err != nil {
		log.Printf("failed to create wiki page %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(page); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// GetById is no longer exposed; lookups go through the slug route instead.
//func (h *WikiPageHandler) GetById(w http.ResponseWriter, r *http.Request) {
//
//	id, err := strconv.Atoi(chi.URLParam(r, "id"))
//	if err != nil {
//		http.Error(w, err.Error(), http.StatusBadRequest)
//		return
//	}
//
//	page, err := h.svc.GetById(r.Context(), int64(id))
//	if err != nil {
//		if errors.Is(err, gorm.ErrRecordNotFound) {
//			http.Error(w, err.Error(), http.StatusNotFound)
//			return
//		}
//		log.Printf("failed to get wiki page %v", err)
//		http.Error(w, "internal server error", http.StatusInternalServerError)
//		return
//	}
//
//	w.Header().Set("Content-Type", "application/json")
//	w.WriteHeader(http.StatusOK)
//	if err := json.NewEncoder(w).Encode(page); err != nil {
//		log.Printf("failed to write response: %v", err)
//	}
//}

func (h *WikiPageHandler) GetBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		http.Error(w, "slug is required", http.StatusBadRequest)
		return
	}

	page, err := h.svc.GetBySlug(r.Context(), slug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		log.Printf("failed to get wiki page by slug %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(page); err != nil {
		log.Printf("failed to write response: %v", err)
	}
}

func (h *WikiPageHandler) Update(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var page model.WikiPageUpsertReq

	if err := json.NewDecoder(r.Body).Decode(&page); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.svc.Update(r.Context(), int64(id), page); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "wiki page not found", http.StatusNotFound)
			return
		}
		log.Printf("failed to update wiki page %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}

func (h *WikiPageHandler) Delete(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.svc.Delete(r.Context(), int64(id)); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "wiki page not found", http.StatusNotFound)
			return
		}
		log.Printf("failed to delete wiki page %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
