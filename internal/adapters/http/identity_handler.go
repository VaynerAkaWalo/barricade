package handlers

import (
	"barricade/internal/domain/identity"
	"barricade/internal/infrastructure/httpserver"
	"encoding/json"
	"net/http"
)

type IdentityResponse struct {
	ID        identity.Id `json:"id"`
	Name      string      `json:"name"`
	CreatedAt int64       `json:"createdAt"`
	UpdatedAt int64       `json:"updatedAt"`
}

type registerRequest struct {
	Name   string `json:"name"`
	Secret string `json:"secret"`
}

type IdentityHttpHandler struct {
	Service identity.Service
}

func (handler *IdentityHttpHandler) RegisterRoutes(router *http.ServeMux) {
	router.HandleFunc("/register", handler.Register)
}

func (handler *IdentityHttpHandler) Register(w http.ResponseWriter, r *http.Request) {
	var request registerRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		httpserver.WriteError(w, http.StatusBadRequest, "body does not satisfy required schema")
	}

	entity, err := handler.Service.Register(r.Context(), request.Name, request.Secret)
	if err != nil {
		httpserver.WriteError(w, http.StatusInternalServerError, "unable to create identity with specified attibutes")
	}

	dto := IdentityResponse{
		ID:        entity.Id,
		Name:      entity.Name,
		UpdatedAt: entity.UpdatedAt,
		CreatedAt: entity.CreatedAt,
	}

	httpserver.WriteResponse(w, http.StatusCreated, dto)
}
