package oauth2

import (
	"encoding/base64"
	"log/slog"
	"net/http"
	"strings"

	"github.com/VaynerAkaWalo/go-toolkit/xhttp"
)

type TokenHttpHandler struct {
	Service *TokenService
}

func (h *TokenHttpHandler) RegisterRoutes(router *xhttp.Router) {
	router.RegisterHandler("POST /v1/oauth2/token", h.Token)
}

func (h *TokenHttpHandler) Token(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		return xhttp.NewError("invalid_request", http.StatusBadRequest)
	}

	clientId, clientSecret := extractClientCredentials(r)

	params := ExchangeTokenParams{
		GrantType:    r.PostFormValue("grant_type"),
		Code:         r.PostFormValue("code"),
		RedirectURI:  r.PostFormValue("redirect_uri"),
		ClientId:     clientId,
		ClientSecret: clientSecret,
		CodeVerifier: r.PostFormValue("code_verifier"),
	}

	result, err := h.Service.Exchange(ctx, params)
	if err != nil {
		slog.ErrorContext(ctx, err.Error())
		return mapTokenError(err)
	}

	xhttp.WriteResponse(w, http.StatusOK, result)
	return nil
}

func extractClientCredentials(r *http.Request) (string, string) {
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Basic ") {
		decoded, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(authHeader, "Basic "))
		if err != nil {
			return "", ""
		}

		parts := strings.SplitN(string(decoded), ":", 2)
		if len(parts) == 2 {
			return parts[0], parts[1]
		}
	}

	return r.PostFormValue("client_id"), r.PostFormValue("client_secret")
}

func mapTokenError(err error) error {
	switch {
	case err == ErrUnsupportedGrantType:
		return xhttp.NewError("unsupported_grant_type", http.StatusBadRequest)
	case err == ErrInvalidClient:
		return xhttp.NewError("invalid_client", http.StatusUnauthorized)
	case err == ErrInvalidCode:
		return xhttp.NewError("invalid_grant", http.StatusBadRequest)
	case err == ErrCodeExpired:
		return xhttp.NewError("invalid_grant", http.StatusBadRequest)
	case err == ErrCodeMismatch:
		return xhttp.NewError("invalid_grant", http.StatusBadRequest)
	case err == ErrMissingCodeVerifier:
		return xhttp.NewError("invalid_grant", http.StatusBadRequest)
	case err == ErrInvalidCodeVerifier:
		return xhttp.NewError("invalid_grant", http.StatusBadRequest)
	default:
		return xhttp.NewError("server_error", http.StatusInternalServerError)
	}
}
