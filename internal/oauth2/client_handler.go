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
}

type registerClientResponse struct {
	ClientId     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
}

type ClientHttpHandler struct {
	ClientService ClientService
}

func (h *ClientHttpHandler) RegisterRoutes(router *xhttp.Router) {
	router.RegisterHandler("POST /v1/oauth2/clients", h.Register)
}

func (h *ClientHttpHandler) Register(w http.ResponseWriter, r *http.Request) error {
	var request registerClientRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		return xhttp.NewError("request does not satisfy required schema", http.StatusBadRequest)
	}

	result, err := h.ClientService.Register(r.Context(), RegisterClientParams{
		Name:        request.Name,
		Domain:      request.Domain,
		RedirectURI: request.RedirectURI,
	})
	if err != nil {
		return mapClientError(r.Context(), err)
	}

	return xhttp.WriteResponse(w, http.StatusCreated, registerClientResponse{
		ClientId:     string(result.Client.Id),
		ClientSecret: string(result.ClientSecret),
	})
}

func mapClientError(ctx context.Context, err error) error {
	switch {
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
