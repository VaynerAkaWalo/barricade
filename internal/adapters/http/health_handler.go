package handlers

import (
	"barricade/internal/domain/healthcheck"
	"barricade/internal/infrastructure/httpserver"
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

	httpserver.WriteResponse(w, http.StatusOK, dto)
}
