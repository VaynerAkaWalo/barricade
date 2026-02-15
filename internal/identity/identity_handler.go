package identity

import (
	"encoding/json"
	"net/http"

	"github.com/VaynerAkaWalo/go-toolkit/xhttp"
)

type IdentityResponse struct {
	ID        Id     `json:"id"`
	Name      string `json:"name"`
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt int64  `json:"updatedAt"`
}

type registerRequest struct {
	Name   string `json:"name"`
	Secret string `json:"secret"`
}

type HttpHandler struct {
	Service Service
}

func (handler *HttpHandler) RegisterRoutes(router *xhttp.Router) {
	router.RegisterHandler("POST /v1/register", handler.Register)
}

func (handler *HttpHandler) Register(w http.ResponseWriter, r *http.Request) error {
	var request registerRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		return xhttp.NewError("request does not satisfy required schema", http.StatusBadRequest)
	}

	entity, err := handler.Service.Register(r.Context(), request.Name, request.Secret)
	if err != nil {
		return xhttp.NewError("unable to create identity with specified attributes", http.StatusInternalServerError)
	}

	dto := IdentityResponse{
		ID:        entity.Id,
		Name:      entity.Name,
		UpdatedAt: entity.UpdatedAt,
		CreatedAt: entity.CreatedAt,
	}

	return xhttp.WriteResponse(w, http.StatusCreated, dto)
}
