package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// ParseJSONBody parses JSON request body into the provided struct
func ParseJSONBody(r *http.Request, v interface{}) error {
	if r.Body == nil {
		return fmt.Errorf("request body is empty")
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields() // Strict parsing

	if err := decoder.Decode(v); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	return nil
}

// GetPathParam extracts a path parameter from the URL
func GetPathParam(r *http.Request, key string) string {
	vars := mux.Vars(r)
	return vars[key]
}

// GetQueryParam extracts a query parameter with a default value
func GetQueryParam(r *http.Request, key, defaultValue string) string {
	value := r.URL.Query().Get(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// GetQueryParamInt extracts an integer query parameter with a default value
func GetQueryParamInt(r *http.Request, key string, defaultValue int) (int, error) {
	value := r.URL.Query().Get(key)
	if value == "" {
		return defaultValue, nil
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue, fmt.Errorf("invalid integer value for %s: %s", key, value)
	}

	return intValue, nil
}

// GetPaginationParams extracts pagination parameters from query string
func GetPaginationParams(r *http.Request) (page, perPage int, err error) {
	page, err = GetQueryParamInt(r, "page", 1)
	if err != nil {
		return 0, 0, err
	}

	if page < 1 {
		page = 1
	}

	perPage, err = GetQueryParamInt(r, "per_page", 50)
	if err != nil {
		return 0, 0, err
	}

	// Limit per_page to reasonable bounds
	if perPage < 1 {
		perPage = 50
	} else if perPage > 1000 {
		perPage = 1000
	}

	return page, perPage, nil
}

// CalculatePagination calculates pagination metadata
func CalculatePagination(page, perPage, total int) *Meta {
	totalPages := (total + perPage - 1) / perPage // Ceiling division

	return &Meta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	}
}

// ValidateRequired checks if required fields are present in a map
func ValidateRequired(data map[string]interface{}, fields ...string) error {
	for _, field := range fields {
		if value, exists := data[field]; !exists || value == nil || value == "" {
			return fmt.Errorf("field '%s' is required", field)
		}
	}
	return nil
}
