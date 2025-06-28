package handlers

import (
	"barricade/internal/domain/healthcheck"
	"barricade/internal/infrastructure/htp"
	"net/http"
)

type HealthResponse struct {
	Status string `json:"status"`
}

type HealthHttpHandler struct {
	Service *healthcheck.Service
}

func (handler *HealthHttpHandler) RegisterRoutes(router *http.ServeMux) {
	router.Handle("/health", htp.HttpHandler(handler.Get))
}

func (handler *HealthHttpHandler) Get(w http.ResponseWriter, r *http.Request) error {
	serviceHealth := handler.Service.IsSystemHealthy()

	dto := HealthResponse{
		Status: serviceHealth.Status,
	}

	return htp.WriteResponse(w, http.StatusOK, dto)
}
