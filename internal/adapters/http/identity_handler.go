package handlers

import (
	"barricade/internal/domain/identity"
	"barricade/internal/infrastructure/htp"
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
	router.Handle("POST /v1/register", htp.HttpHandler(handler.Register))
}

func (handler *IdentityHttpHandler) Register(w http.ResponseWriter, r *http.Request) error {
	var request registerRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		return htp.NewError("request does not satisfy required schema", http.StatusBadRequest)
	}

	entity, err := handler.Service.Register(r.Context(), request.Name, request.Secret)
	if err != nil {
		return htp.NewError("unable to create identity with specified attributes", http.StatusInternalServerError)
	}

	dto := IdentityResponse{
		ID:        entity.Id,
		Name:      entity.Name,
		UpdatedAt: entity.UpdatedAt,
		CreatedAt: entity.CreatedAt,
	}

	return htp.WriteResponse(w, http.StatusCreated, dto)
}
