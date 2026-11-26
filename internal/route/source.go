package route

import (
	"encoding/json"
	"errors"
	"llm-wiki-server/internal/service"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type SourceHandler struct {
	svc service.ISourceService
}

func NewSourceHandler(svc service.ISourceService) *SourceHandler {
	return &SourceHandler{svc: svc}
}

func (h *SourceHandler) RegisterRoutes(router chi.Router) {
	router.Route("/api/sources", func(r chi.Router) {
		r.Get("/", h.List)
		r.Post("/", h.Upsert)
		r.Get("/{id}", h.Get)
		r.Delete("/{id}", h.Delete)
		r.Get("/{id}/pages", h.ListPages)
	})
}

type upsertSourceRequest struct {
	Origin  string `json:"origin"`
	Title   string `json:"title"`
	Summary string `json:"summary"`
}

func (h *SourceHandler) List(w http.ResponseWriter, r *http.Request) {
	keyword := r.URL.Query().Get("keyword")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	size, _ := strconv.Atoi(r.URL.Query().Get("size"))

	resp, err := h.svc.List(r.Context(), keyword, page, size)
	if err != nil {
		log.Printf("failed to list sources: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("failed to write response: %v", err)
	}
}

func (h *SourceHandler) Upsert(w http.ResponseWriter, r *http.Request) {
	var req upsertSourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.Origin == "" {
		http.Error(w, "origin is required", http.StatusBadRequest)
		return
	}

	source, err := h.svc.UpsertByOrigin(r.Context(), req.Origin, req.Title, req.Summary)
	if err != nil {
		log.Printf("failed to upsert source %q: %v", req.Origin, err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(source); err != nil {
		log.Printf("failed to write response: %v", err)
	}
}

func (h *SourceHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid source id", http.StatusBadRequest)
		return
	}

	source, err := h.svc.Get(r.Context(), int64(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "source not found", http.StatusNotFound)
			return
		}
		log.Printf("failed to get source %d: %v", id, err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(source); err != nil {
		log.Printf("failed to write response: %v", err)
	}
}

func (h *SourceHandler) ListPages(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid source id", http.StatusBadRequest)
		return
	}

	pages, err := h.svc.ListPagesBySource(r.Context(), int64(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "source not found", http.StatusNotFound)
			return
		}
		log.Printf("failed to list pages for source %d: %v", id, err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(pages); err != nil {
		log.Printf("failed to write response: %v", err)
	}
}

func (h *SourceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid source id", http.StatusBadRequest)
		return
	}

	if err := h.svc.Delete(r.Context(), int64(id)); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "source not found", http.StatusNotFound)
			return
		}
		log.Printf("failed to delete source %d: %v", id, err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
