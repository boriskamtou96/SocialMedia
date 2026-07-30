package main

import (
	"encoding/json"
	"net/http"

	validator2 "github.com/go-playground/validator/v10"
)

var Validate *validator2.Validate

func init() {
	Validate = validator2.New(validator2.WithRequiredStructEnabled())
}

func WriteJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	encoder := json.NewEncoder(w)

	return encoder.Encode(data)
}

func ReadJSON(w http.ResponseWriter, r *http.Request, data any) error {
	maxBytes := 1_048_576 // 1MB
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	w.Header().Set("Content-Type", "application/json")

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	return decoder.Decode(data)
}

func ErrorJSON(w http.ResponseWriter, status int, errorCode string, message string) {
	type ErrorResponse struct {
		Status    int    `json:"status"`
		ErrorCode string `json:"errorCode"`
		Message   string `json:"message"`
	}
	errorResponse := ErrorResponse{
		Status:    status,
		ErrorCode: errorCode,
		Message:   message,
	}
	if err := WriteJSON(w, status, errorResponse); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
