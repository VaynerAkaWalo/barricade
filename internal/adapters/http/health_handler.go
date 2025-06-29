package handlers

import (
	"barricade/internal/infrastructure/htp"
	"net/http"
)

type HealthResponse struct {
	Status string `json:"status"`
}

type HealthHttpHandler struct {
}

func (handler *HealthHttpHandler) RegisterRoutes(router *http.ServeMux) {
	router.Handle("/health", htp.HttpHandler(handler.Get))
}

func (handler *HealthHttpHandler) Get(w http.ResponseWriter, r *http.Request) error {
	dto := HealthResponse{
		Status: "ok",
	}

	return htp.WriteResponse(w, http.StatusOK, dto)
}
