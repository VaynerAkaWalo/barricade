package httpserver

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func WriteResponse(w http.ResponseWriter, code int, response interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Internal server error")
	}
}

func WriteError(w http.ResponseWriter, code int, message string) {
	WriteResponse(w, code, ErrorResponse{Message: message, Code: code})
}
