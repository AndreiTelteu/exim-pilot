package api

import (
	"encoding/json"
	"net/http"
)

// APIResponse represents the standard API response structure
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// Meta contains pagination and additional metadata
type Meta struct {
	Page       int `json:"page,omitempty"`
	PerPage    int `json:"per_page,omitempty"`
	Total      int `json:"total,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`
}

// WriteJSONResponse writes a standardized JSON response
func WriteJSONResponse(w http.ResponseWriter, statusCode int, response APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		// If we can't encode the response, write a basic error
		http.Error(w, `{"success":false,"error":"Failed to encode response"}`, http.StatusInternalServerError)
	}
}

// WriteSuccessResponse writes a successful response with data
func WriteSuccessResponse(w http.ResponseWriter, data interface{}) {
	response := APIResponse{
		Success: true,
		Data:    data,
	}
	WriteJSONResponse(w, http.StatusOK, response)
}

// WriteSuccessResponseWithMeta writes a successful response with data and metadata
func WriteSuccessResponseWithMeta(w http.ResponseWriter, data interface{}, meta *Meta) {
	response := APIResponse{
		Success: true,
		Data:    data,
		Meta:    meta,
	}
	WriteJSONResponse(w, http.StatusOK, response)
}

// WriteErrorResponse writes an error response
func WriteErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	response := APIResponse{
		Success: false,
		Error:   message,
	}
	WriteJSONResponse(w, statusCode, response)
}

// WriteBadRequestResponse writes a 400 Bad Request response
func WriteBadRequestResponse(w http.ResponseWriter, message string) {
	WriteErrorResponse(w, http.StatusBadRequest, message)
}

// WriteNotFoundResponse writes a 404 Not Found response
func WriteNotFoundResponse(w http.ResponseWriter, message string) {
	WriteErrorResponse(w, http.StatusNotFound, message)
}

// WriteInternalErrorResponse writes a 500 Internal Server Error response
func WriteInternalErrorResponse(w http.ResponseWriter, message string) {
	WriteErrorResponse(w, http.StatusInternalServerError, message)
}

// WriteUnauthorizedResponse writes a 401 Unauthorized response
func WriteUnauthorizedResponse(w http.ResponseWriter, message string) {
	WriteErrorResponse(w, http.StatusUnauthorized, message)
}

// WriteForbiddenResponse writes a 403 Forbidden response
func WriteForbiddenResponse(w http.ResponseWriter, message string) {
	WriteErrorResponse(w, http.StatusForbidden, message)
}
