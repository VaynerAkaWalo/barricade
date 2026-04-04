package oauth2

import (
	"barricade/internal/authentication"
	"barricade/internal/identity"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/VaynerAkaWalo/go-toolkit/xhttp"
)

type HttpHandler struct {
	Service            *AuthorizeService
	AuthService        *authentication.Service
	LoginURL           string
	DefaultRedirectURI string
}

func (h *HttpHandler) RegisterRoutes(router *xhttp.Router) {
	router.RegisterHandler("GET /v1/oauth2/authorize", h.Authorize)
}

func (h *HttpHandler) Authorize(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	params := AuthorizationParams{
		ResponseType: r.URL.Query().Get("response_type"),
		ClientId:     r.URL.Query().Get("client_id"),
		Scope:        r.URL.Query().Get("scope"),
	}

	redirectURI := r.URL.Query().Get("redirect_uri")
	if redirectURI == "" {
		redirectURI = h.DefaultRedirectURI
	}

	if err := validateRedirectURI(redirectURI); err != nil {
		redirectURL := buildErrorRedirectURL(redirectURI, mapErrorToCode(err), err.Error())
		http.Redirect(w, r, redirectURL, http.StatusFound)
		return nil
	}

	if err := h.Service.Validate(params); err != nil {
		redirectURL := buildErrorRedirectURL(redirectURI, mapErrorToCode(err), err.Error())
		http.Redirect(w, r, redirectURL, http.StatusFound)
		return nil
	}

	identityId, err := h.authenticate(ctx, r)
	if err != nil {
		slog.ErrorContext(ctx, "authentication failed", "error", err)
		loginRedirect := h.buildLoginRedirectURL(r)
		http.Redirect(w, r, loginRedirect, http.StatusFound)
		return nil
	}

	result, err := h.Service.Authorize(ctx, identity.Id(identityId), params.ClientId)
	if err != nil {
		redirectURL := buildErrorRedirectURL(redirectURI, "server_error", "failed to generate token")
		http.Redirect(w, r, redirectURL, http.StatusFound)
		return nil
	}

	redirectURL := buildSuccessRedirectURL(redirectURI, result, h.Service.TokenExpiry)
	http.Redirect(w, r, redirectURL, http.StatusFound)
	return nil
}

func (h *HttpHandler) authenticate(ctx context.Context, r *http.Request) (string, error) {
	if identityId, ok := ctx.Value(xhttp.UserId).(string); ok && identityId != "" {
		return identityId, nil
	}

	cookie, err := r.Cookie(authentication.SessionCookie)
	if err != nil {
		return "", err
	}

	ident, err := h.AuthService.AuthenticateBySession(ctx, authentication.SessionId(cookie.Value))
	if err != nil {
		return "", err
	}

	return string(ident.Id), nil
}

func (h *HttpHandler) buildLoginRedirectURL(r *http.Request) string {
	u, err := url.Parse(h.LoginURL)
	if err != nil {
		return h.LoginURL
	}
	q := u.Query()
	q.Set("target", r.URL.String())
	u.RawQuery = q.Encode()
	return u.String()
}

func validateRedirectURI(redirectURI string) error {
	if redirectURI == "" {
		return ErrInvalidRequest
	}
	if _, err := url.ParseRequestURI(redirectURI); err != nil {
		return ErrInvalidRedirectURI
	}
	return nil
}

func buildSuccessRedirectURL(redirectURI string, result *AuthorizationResult, tokenExpiry int) string {
	u, err := url.Parse(redirectURI)
	if err != nil {
		return redirectURI
	}

	fragment := url.Values{}
	fragment.Set("id_token", string(result.IDToken))
	fragment.Set("token_type", "Bearer")
	fragment.Set("expires_in", fmt.Sprintf("%d", tokenExpiry*60))

	u.Fragment = fragment.Encode()

	return u.String()
}

func buildErrorRedirectURL(redirectURI string, errorCode string, errorDescription string) string {
	u, err := url.Parse(redirectURI)
	if err != nil {
		return redirectURI
	}

	fragment := url.Values{}
	fragment.Set("error", errorCode)
	if errorDescription != "" {
		fragment.Set("error_description", errorDescription)
	}

	u.Fragment = fragment.Encode()

	return u.String()
}

func mapErrorToCode(err error) string {
	switch {
	case errors.Is(err, ErrInvalidRequest):
		return "invalid_request"
	case errors.Is(err, ErrUnsupportedResponseType):
		return "unsupported_response_type"
	case errors.Is(err, ErrInvalidScope):
		return "invalid_scope"
	case errors.Is(err, ErrInvalidRedirectURI):
		return "invalid_request"
	default:
		return "server_error"
	}
}
