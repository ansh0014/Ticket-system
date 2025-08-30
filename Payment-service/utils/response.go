package utils

import (
    "encoding/json"
    "net/http"
    "time"
)

// Response is the standard API response structure
type Response struct {
    Success    bool        `json:"success"`
    Message    string      `json:"message,omitempty"`
    Data       interface{} `json:"data,omitempty"`
    Errors     interface{} `json:"errors,omitempty"`
    StatusCode int         `json:"status_code"`
    Timestamp  time.Time   `json:"timestamp"`
}

// ValidationError represents field validation errors
type ValidationError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
}

// ErrorResponse sends a standard error response
func ErrorResponse(w http.ResponseWriter, message string, statusCode int, errors interface{}) {
    response := Response{
        Success:    false,
        Message:    message,
        Errors:     errors,
        StatusCode: statusCode,
        Timestamp:  time.Now(),
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    json.NewEncoder(w).Encode(response)
}

// SuccessResponse sends a standard success response
func SuccessResponse(w http.ResponseWriter, message string, data interface{}, statusCode int) {
    response := Response{
        Success:    true,
        Message:    message,
        Data:       data,
        StatusCode: statusCode,
        Timestamp:  time.Now(),
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    json.NewEncoder(w).Encode(response)
}

// CreatedResponse is a shorthand for 201 Created responses
func CreatedResponse(w http.ResponseWriter, message string, data interface{}) {
    SuccessResponse(w, message, data, http.StatusCreated)
}

// OkResponse is a shorthand for 200 OK responses
func OkResponse(w http.ResponseWriter, message string, data interface{}) {
    SuccessResponse(w, message, data, http.StatusOK)
}

// BadRequestResponse is a shorthand for 400 Bad Request responses
func BadRequestResponse(w http.ResponseWriter, message string, errors interface{}) {
    ErrorResponse(w, message, http.StatusBadRequest, errors)
}

// UnauthorizedResponse is a shorthand for 401 Unauthorized responses
func UnauthorizedResponse(w http.ResponseWriter, message string) {
    ErrorResponse(w, message, http.StatusUnauthorized, nil)
}

// ForbiddenResponse is a shorthand for 403 Forbidden responses
func ForbiddenResponse(w http.ResponseWriter, message string) {
    ErrorResponse(w, message, http.StatusForbidden, nil)
}

// NotFoundResponse is a shorthand for 404 Not Found responses
func NotFoundResponse(w http.ResponseWriter, message string) {
    ErrorResponse(w, message, http.StatusNotFound, nil)
}

// ConflictResponse is a shorthand for 409 Conflict responses
func ConflictResponse(w http.ResponseWriter, message string) {
    ErrorResponse(w, message, http.StatusConflict, nil)
}

// ServerErrorResponse is a shorthand for 500 Internal Server Error responses
func ServerErrorResponse(w http.ResponseWriter, message string) {
    ErrorResponse(w, message, http.StatusInternalServerError, nil)
}

// ValidationErrorResponse creates a response for validation errors
func ValidationErrorResponse(w http.ResponseWriter, errors []ValidationError) {
    ErrorResponse(w, "Validation failed", http.StatusBadRequest, errors)
}