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

type whoAmIResponse struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type AuthenticationHttpHandler struct {
	Service authentication.SessionService
	Domain  string
}

func (handler *AuthenticationHttpHandler) RegisterRoutes(router *xhttp.Router) {
	router.RegisterHandler("POST /v1/login", handler.login)
	router.RegisterHandler("POST /v1/logout", handler.logout)
	router.RegisterHandler("GET /v1/whoami", handler.whoAmI)
}

func (handler *AuthenticationHttpHandler) login(w http.ResponseWriter, r *http.Request) error {
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
		Domain:   "." + handler.Domain,
		MaxAge:   300,
		HttpOnly: true,
		Secure:   true,
	}
	http.SetCookie(w, &sessionCookie)
	return xhttp.WriteResponse(w, http.StatusNoContent, "")
}

func (handler *AuthenticationHttpHandler) logout(w http.ResponseWriter, r *http.Request) error {
	_, err := r.Cookie(SessionCookie)
	if err != nil {
		return xhttp.WriteResponse(w, http.StatusOK, struct{}{})
	}

	cookie := http.Cookie{
		Name:     SessionCookie,
		Value:    "",
		Domain:   "." + handler.Domain,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
	}

	http.SetCookie(w, &cookie)
	return xhttp.WriteResponse(w, http.StatusAccepted, "")
}

func (handler *AuthenticationHttpHandler) whoAmI(w http.ResponseWriter, r *http.Request) error {
	sessionCookie, err := r.Cookie(SessionCookie)
	if err != nil {
		return xhttp.NewError("missing session in request", http.StatusUnauthorized)
	}

	identity, err := handler.Service.GetIdentityBySession(r.Context(), authentication.SessionId(sessionCookie.Value))
	if err != nil {
		return err
	}

	response := whoAmIResponse{
		Id:   string(identity.Id),
		Name: identity.Name,
	}

	return xhttp.WriteResponse(w, http.StatusOK, response)
}
