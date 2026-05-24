package response

import (
	"encoding/json"
	"net/http"
)

type SuccessResponse struct {
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type ErrorResponse struct {
	Message string              `json:"message"`
	Errors  map[string][]string `json:"errors,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func WriteOK(w http.ResponseWriter, message string, data any) {
	writeJSON(w, http.StatusOK, SuccessResponse{
		Message: message,
		Data:    data,
	})
}

func WriteCreated(w http.ResponseWriter, message string, data any) {
	writeJSON(w, http.StatusCreated, SuccessResponse{
		Message: message,
		Data:    data,
	})
}

func WriteNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func WriteError(w http.ResponseWriter, status int, message string, errs map[string][]string) {
	writeJSON(w, status, ErrorResponse{
		Message: message,
		Errors:  errs,
	})
}

func WriteBadRequest(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusBadRequest, message, nil)
}

func WriteNotFound(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusNotFound, message, nil)
}

func WriteConflict(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusConflict, message, nil)
}

func WriteValidationFailed(w http.ResponseWriter, message string, errs map[string][]string) {
	WriteError(w, http.StatusUnprocessableEntity, message, errs)
}

func WriteInternalServerError(w http.ResponseWriter) {
	WriteError(w, http.StatusInternalServerError, "internal server error", nil)
}

func WriteTooManyRequests(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusTooManyRequests, message, nil)
}
