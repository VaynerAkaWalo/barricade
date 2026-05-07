package oauth2

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/VaynerAkaWalo/go-toolkit/xhttp"
)

type registerClientRequest struct {
	Name        string `json:"name"`
	Domain      string `json:"domain"`
	RedirectURI string `json:"redirectURI"`
	ClientType  string `json:"clientType"`
}

type registerClientResponse struct {
	ClientId     string `json:"clientId"`
	ClientSecret string `json:"clientSecret,omitempty"`
}

type listClientResponse struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Domain      string `json:"domain"`
	RedirectURI string `json:"redirectURI"`
	Type        string `json:"type"`
	CreatedAt   int64  `json:"createdAt"`
	UpdatedAt   int64  `json:"updatedAt"`
}

type ClientHttpHandler struct {
	ClientService ClientService
}

func (h *ClientHttpHandler) RegisterRoutes(router *xhttp.Router) {
	router.RegisterHandler("POST /v1/oauth2/clients", h.Register)
	router.RegisterHandler("GET /v1/oauth2/clients", h.List)
}

func (h *ClientHttpHandler) Register(w http.ResponseWriter, r *http.Request) error {
	ownerId, ok := r.Context().Value(xhttp.UserId).(string)
	if !ok || ownerId == "" {
		return xhttp.NewError("unauthorized", http.StatusUnauthorized)
	}

	var request registerClientRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		return xhttp.NewError("request does not satisfy required schema", http.StatusBadRequest)
	}

	result, err := h.ClientService.Register(r.Context(), RegisterClientParams{
		OwnerId:     ownerId,
		Name:        request.Name,
		Domain:      request.Domain,
		RedirectURI: request.RedirectURI,
		ClientType:  ClientType(request.ClientType),
	})
	if err != nil {
		return mapClientError(r.Context(), err)
	}

	return xhttp.WriteResponse(w, http.StatusCreated, registerClientResponse{
		ClientId:     string(result.Client.Id),
		ClientSecret: string(result.ClientSecret),
	})
}

func (h *ClientHttpHandler) List(w http.ResponseWriter, r *http.Request) error {
	ownerId, ok := r.Context().Value(xhttp.UserId).(string)
	if !ok || ownerId == "" {
		return xhttp.NewError("unauthorized", http.StatusUnauthorized)
	}

	clients, err := h.ClientService.FindAll(r.Context())
	if err != nil {
		return mapClientError(r.Context(), err)
	}

	response := make([]listClientResponse, 0)
	for _, c := range clients {
		if c.OwnerId != ownerId {
			continue
		}
		response = append(response, listClientResponse{
			Id:          string(c.Id),
			Name:        c.Name,
			Domain:      c.Domain,
			RedirectURI: c.RedirectURI,
			Type:        string(c.Type),
			CreatedAt:   c.CreatedAt,
			UpdatedAt:   c.UpdatedAt,
		})
	}

	return xhttp.WriteResponse(w, http.StatusOK, response)
}

func mapClientError(ctx context.Context, err error) error {
	switch {
	case errors.Is(err, ErrInvalidClientType):
		return xhttp.NewError("client type must be 'public' or 'confidential'", http.StatusBadRequest)
	case errors.Is(err, ErrClientEmptyOwnerId):
		return xhttp.NewError("unauthorized", http.StatusUnauthorized)
	case errors.Is(err, ErrClientEmptyName):
		return xhttp.NewError("name is required", http.StatusBadRequest)
	case errors.Is(err, ErrClientEmptyDomain):
		return xhttp.NewError("domain is required", http.StatusBadRequest)
	case errors.Is(err, ErrClientEmptyRedirectURI):
		return xhttp.NewError("redirectURI is required", http.StatusBadRequest)
	case errors.Is(err, ErrClientInvalidRedirectURI):
		return xhttp.NewError("redirectURI is not a valid URL", http.StatusBadRequest)
	case errors.Is(err, ErrClientRedirectURIDomainMismatch):
		return xhttp.NewError("redirectURI domain does not match client domain", http.StatusBadRequest)
	case errors.Is(err, ErrClientNotFound):
		return xhttp.NewError("client not found", http.StatusNotFound)
	default:
		slog.ErrorContext(ctx, "client service error", "error", err)
		return xhttp.NewError("unable to register client", http.StatusInternalServerError)
	}
}
