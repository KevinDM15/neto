package handler

import (
	"net/http"
)

// HealthHandler maneja el endpoint de health check.
type HealthHandler struct{}

// NewHealthHandler crea un nuevo HealthHandler.
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Health responde con el estado del servidor.
// GET /health
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"version": "0.1.0",
	})
}
