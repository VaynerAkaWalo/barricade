package handlers

import (
	"barricade/internal/domain/healthcheck"
	"encoding/json"
	"net/http"
)

type HealthResponse struct {
	Status string `json:"status"`
}

type HealthHttpHandler struct {
	Service *healthcheck.Service
}

func (handler *HealthHttpHandler) RegisterRoutes(router *http.ServeMux) {
	router.HandleFunc("/health", handler.Get)
}

func (handler *HealthHttpHandler) Get(w http.ResponseWriter, r *http.Request) {
	serviceHealth := handler.Service.IsSystemHealthy()

	dto := HealthResponse{
		Status: serviceHealth.Status,
	}

	w.Header().Set("Content-Type", "application/json")

	err := json.NewEncoder(w).Encode(dto)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
