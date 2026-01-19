package helpers

import (
	"encoding/json"
	"net/http"
)

// JSON writes a JSON response with the given status code
func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

// JSONOk writes a 200 OK JSON response
func JSONOk(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusOK, data)
}

// JSONCreated writes a 201 Created JSON response
func JSONCreated(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusCreated, data)
}

// JSONNoContent writes a 204 No Content response
func JSONNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// ErrorResponse represents an error response body
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    string `json:"code,omitempty"`
}

// JSONError writes a JSON error response
func JSONError(w http.ResponseWriter, status int, message string) {
	JSON(w, status, ErrorResponse{
		Error:   http.StatusText(status),
		Message: message,
	})
}

// JSONBadRequest writes a 400 Bad Request error
func JSONBadRequest(w http.ResponseWriter, message string) {
	JSONError(w, http.StatusBadRequest, message)
}

// JSONUnauthorized writes a 401 Unauthorized error
func JSONUnauthorized(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Unauthorized"
	}
	JSONError(w, http.StatusUnauthorized, message)
}

// JSONForbidden writes a 403 Forbidden error
func JSONForbidden(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Forbidden"
	}
	JSONError(w, http.StatusForbidden, message)
}

// JSONNotFound writes a 404 Not Found error
func JSONNotFound(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Resource not found"
	}
	JSONError(w, http.StatusNotFound, message)
}

// JSONInternalError writes a 500 Internal Server Error
func JSONInternalError(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Internal server error"
	}
	JSONError(w, http.StatusInternalServerError, message)
}

// DecodeJSON decodes a JSON request body into the given struct
func DecodeJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

// Redirect sends an HTTP redirect response
func Redirect(w http.ResponseWriter, r *http.Request, url string, status int) {
	http.Redirect(w, r, url, status)
}
