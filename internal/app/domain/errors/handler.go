package errors

import (
	"net/http"

	errorpages "github.com/FACorreiaa/talentsynapse/internal/app/views/pages/errors"
)

// Handler handles error page responses
type Handler struct{}

// NewHandler creates a new error handler
func NewHandler() *Handler {
	return &Handler{}
}

// NotFound renders the 404 page
func (h *Handler) NotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	component := errorpages.NotFound()
	_ = component.Render(r.Context(), w)
}

// ServerError renders the 500 page with an optional message
func (h *Handler) ServerError(w http.ResponseWriter, r *http.Request, message string) {
	w.WriteHeader(http.StatusInternalServerError)
	component := errorpages.ServerError(message)
	_ = component.Render(r.Context(), w)
}

// Forbidden renders the 403 page
func (h *Handler) Forbidden(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusForbidden)
	component := errorpages.Forbidden()
	_ = component.Render(r.Context(), w)
}

// RenderNotFound is a helper function to render 404 without handler instance
func RenderNotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	component := errorpages.NotFound()
	_ = component.Render(r.Context(), w)
}

// RenderServerError is a helper function to render 500 without handler instance
func RenderServerError(w http.ResponseWriter, r *http.Request, message string) {
	w.WriteHeader(http.StatusInternalServerError)
	component := errorpages.ServerError(message)
	_ = component.Render(r.Context(), w)
}
