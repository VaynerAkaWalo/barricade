package handlers

import (
	"github.com/VaynerAkaWalo/go-toolkit/xhttp"
	"net/http"
)

type HealthResponse struct {
	Status string `json:"status"`
}

type HealthHttpHandler struct {
}

func (handler *HealthHttpHandler) RegisterRoutes(router *xhttp.Router) {
	router.RegisterHandler("/health", handler.Get)
}

func (handler *HealthHttpHandler) Get(w http.ResponseWriter, r *http.Request) error {
	dto := HealthResponse{
		Status: "ok",
	}

	return xhttp.WriteResponse(w, http.StatusOK, dto)
}
