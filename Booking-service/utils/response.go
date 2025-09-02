package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Response represents a standard API response structure
type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// RespondWithJSON writes a JSON response with the given status code
func RespondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(payload)
}

// RespondWithError writes an error response with the given status code and message
func RespondWithError(w http.ResponseWriter, statusCode int, message string) {
	response := Response{
		Status:  "error",
		Message: message,
	}
	RespondWithJSON(w, statusCode, response)
}

// OkResponse sends a successful response with status 200
func OkResponse(w http.ResponseWriter, message string, data interface{}) {
	response := Response{
		Status:  "success",
		Message: message,
		Data:    data,
	}
	RespondWithJSON(w, http.StatusOK, response)
}

// CreatedResponse sends a successful response with status 201
func CreatedResponse(w http.ResponseWriter, message string, data interface{}) {
	response := Response{
		Status:  "success",
		Message: message,
		Data:    data,
	}
	RespondWithJSON(w, http.StatusCreated, response)
}

// BadRequestResponse sends a 400 Bad Request response
func BadRequestResponse(w http.ResponseWriter, message string, errors []ValidationError) {
	response := Response{
		Status:  "error",
		Message: message,
		Error:   errors,
	}
	RespondWithJSON(w, http.StatusBadRequest, response)
}

// UnauthorizedResponse sends a 401 Unauthorized response
func UnauthorizedResponse(w http.ResponseWriter, message string) {
	response := Response{
		Status:  "error",
		Message: message,
	}
	RespondWithJSON(w, http.StatusUnauthorized, response)
}

// ForbiddenResponse sends a 403 Forbidden response
func ForbiddenResponse(w http.ResponseWriter, message string) {
	response := Response{
		Status:  "error",
		Message: message,
	}
	RespondWithJSON(w, http.StatusForbidden, response)
}

// NotFoundResponse sends a 404 Not Found response
func NotFoundResponse(w http.ResponseWriter, message string) {
	response := Response{
		Status:  "error",
		Message: message,
	}
	RespondWithJSON(w, http.StatusNotFound, response)
}

// ConflictResponse sends a 409 Conflict response
func ConflictResponse(w http.ResponseWriter, message string) {
	response := Response{
		Status:  "error",
		Message: message,
	}
	RespondWithJSON(w, http.StatusConflict, response)
}

// ServerErrorResponse sends a 500 Internal Server Error response
func ServerErrorResponse(w http.ResponseWriter, message string) {
	response := Response{
		Status:  "error",
		Message: message,
	}
	RespondWithJSON(w, http.StatusInternalServerError, response)
}

// ValidationErrorResponse creates a response for validation errors
func ValidationErrorResponse(w http.ResponseWriter, errors []ValidationError) {
	response := Response{
		Status:  "error",
		Message: "Validation failed",
		Error:   errors,
	}
	RespondWithJSON(w, http.StatusBadRequest, response)
}

// ReadJSON reads the request body into the provided destination struct
func ReadJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	// Limit the size of the request body to 1MB
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	// Initialize the json.Decoder
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	// Decode the request body into the target destination
	err := dec.Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains malformed JSON (at character %d)", syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains malformed JSON")

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown field %s", fieldName)

		case err.Error() == "http: request body too large":
			return fmt.Errorf("body must not be larger than %d bytes", maxBytes)

		case errors.As(err, &invalidUnmarshalError):
			return fmt.Errorf("internal server error: %s", err.Error())

		default:
			return err
		}
	}

	// Check for any additional data in the request body
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single JSON object")
	}

	return nil
}

// GetPathParam extracts a path parameter from the request
func GetPathParam(r *http.Request, name string) string {
	vars := mux.Vars(r)
	return vars[name]
}

// GetObjectID extracts a path parameter and converts it to a MongoDB ObjectID
func GetObjectID(r *http.Request, name string) (primitive.ObjectID, error) {
	id := GetPathParam(r, name)
	if id == "" {
		return primitive.ObjectID{}, fmt.Errorf("missing %s parameter", name)
	}

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return primitive.ObjectID{}, fmt.Errorf("invalid %s format", name)
	}

	return objID, nil
}

// GetQueryParam extracts a query parameter from the request
func GetQueryParam(r *http.Request, name string) string {
	return r.URL.Query().Get(name)
}

// GetQueryParamInt extracts a query parameter and converts it to an integer
func GetQueryParamInt(r *http.Request, name string, defaultValue int) int {
	strValue := GetQueryParam(r, name)
	if strValue == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(strValue)
	if err != nil {
		return defaultValue
	}

	return value
}

// GetQueryParamBool extracts a query parameter and converts it to a boolean
func GetQueryParamBool(r *http.Request, name string, defaultValue bool) bool {
	strValue := GetQueryParam(r, name)
	if strValue == "" {
		return defaultValue
	}

	value, err := strconv.ParseBool(strValue)
	if err != nil {
		return defaultValue
	}

	return value
}

// GetQueryParamDate extracts a query parameter and converts it to a time.Time
func GetQueryParamDate(r *http.Request, name string, layout string) (time.Time, error) {
	strValue := GetQueryParam(r, name)
	if strValue == "" {
		return time.Time{}, nil
	}

	date, err := time.Parse(layout, strValue)
	if err != nil {
		return time.Time{}, err
	}

	return date, nil
}

// GetPageParams extracts pagination parameters from the request
func GetPageParams(r *http.Request) (page, pageSize int) {
	page = GetQueryParamInt(r, "page", 1)
	if page < 1 {
		page = 1
	}

	pageSize = GetQueryParamInt(r, "page_size", 10)
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	return page, pageSize
}
