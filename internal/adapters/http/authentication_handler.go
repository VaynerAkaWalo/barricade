package handlers

import (
	"barricade/internal/domain/authentication"
	"encoding/json"
	"github.com/VaynerAkaWalo/go-toolkit/xhttp"
	"net/http"
)

const SessionCookie = "session_id"

type loginRequest struct {
	Name   string `json:"name"`
	Secret string `json:"secret"`
}

type sessionResponse struct {
	SessionID string `json:"sessionID"`
	Owner     string `json:"owner"`
}

type AuthenticationHttpHandler struct {
	Service authentication.SessionService
}

func (handler *AuthenticationHttpHandler) RegisterRoutes(router *xhttp.Router) {
	router.RegisterHandler("POST /v1/login", handler.Login)
	router.RegisterHandler("POST /v1/auth/session", handler.AuthBySession)
}

func (handler *AuthenticationHttpHandler) Login(w http.ResponseWriter, r *http.Request) error {
	var request loginRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		return xhttp.NewError("request does not satisfy required schema", http.StatusBadRequest)
	}

	session, err := handler.Service.Login(r.Context(), request.Name, request.Secret)
	if err != nil {
		return err
	}

	sessionCookie := http.Cookie{
		Name:     SessionCookie,
		Value:    string(session.Id),
		MaxAge:   300,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
	}
	http.SetCookie(w, &sessionCookie)
	return xhttp.WriteResponse(w, http.StatusOK, struct{}{})
}

func (handler *AuthenticationHttpHandler) AuthBySession(w http.ResponseWriter, r *http.Request) error {
	sessionCookie, err := r.Cookie(SessionCookie)
	if err != nil {
		return xhttp.NewError("missing session in request", http.StatusUnauthorized)
	}

	session, err := handler.Service.AuthenticateBySession(r.Context(), authentication.SessionId(sessionCookie.Value))
	if err != nil {
		return err
	}

	response := sessionResponse{
		SessionID: string(session.Id),
		Owner:     string(session.Owner),
	}

	return xhttp.WriteResponse(w, http.StatusOK, response)
}
