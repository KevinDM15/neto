package handler

import (
	"context"
	"log"
	"net/http"

	"github.com/google/uuid"

	"github.com/neto-app/neto/api/internal/domain/entity"
	"github.com/neto-app/neto/api/internal/middleware"
)

// categoryUseCase define la interfaz que necesita CategoryHandler.
type categoryUseCase interface {
	CreateCategory(ctx context.Context, userID uuid.UUID, name string, parentID *uuid.UUID) (entity.Category, error)
	ListCategories(ctx context.Context, userID uuid.UUID) ([]entity.Category, error)
}

// CategoryHandler maneja los endpoints HTTP de categorías.
type CategoryHandler struct {
	uc categoryUseCase
}

// NewCategoryHandler crea un nuevo CategoryHandler.
func NewCategoryHandler(uc categoryUseCase) *CategoryHandler {
	return &CategoryHandler{uc: uc}
}

// createCategoryRequest es el body esperado para crear una categoría.
type createCategoryRequest struct {
	Name     string  `json:"name"`
	ParentID *string `json:"parent_id,omitempty"`
}

// categoryResponse es la representación JSON de una categoría.
type categoryResponse struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	ParentID  *string `json:"parent_id,omitempty"`
	Icon      string  `json:"icon,omitempty"`
	CreatedAt string  `json:"created_at"`
}

func toCategoryResponse(c entity.Category) categoryResponse {
	resp := categoryResponse{
		ID:        c.ID.String(),
		Name:      c.Name,
		Icon:      c.Icon,
		CreatedAt: c.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if c.ParentID != nil {
		s := c.ParentID.String()
		resp.ParentID = &s
	}
	return resp
}

// Create crea una nueva categoría para el usuario autenticado.
// POST /api/v1/categories
func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req createCategoryRequest
	if err := decode(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	var parentID *uuid.UUID
	if req.ParentID != nil {
		id, err := uuid.Parse(*req.ParentID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid parent_id")
			return
		}
		parentID = &id
	}

	category, err := h.uc.CreateCategory(r.Context(), userID, req.Name, parentID)
	if err != nil {
		log.Printf("handler: create category: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to create category")
		return
	}

	writeJSON(w, http.StatusCreated, toCategoryResponse(category))
}

// List retorna todas las categorías del usuario autenticado.
// GET /api/v1/categories
func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	categories, err := h.uc.ListCategories(r.Context(), userID)
	if err != nil {
		log.Printf("handler: list categories: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to list categories")
		return
	}

	resp := make([]categoryResponse, 0, len(categories))
	for _, c := range categories {
		resp = append(resp, toCategoryResponse(c))
	}

	writeJSON(w, http.StatusOK, resp)
}
