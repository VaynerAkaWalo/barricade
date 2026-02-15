package authentication

import (
	"barricade/internal/identity"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/VaynerAkaWalo/go-toolkit/xhttp"
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

type HttpHandler struct {
	Service     SessionService
	Domain      string
	SessionTime int
}

func (handler *HttpHandler) RegisterRoutes(router *xhttp.Router) {
	router.RegisterHandler("POST /v1/login", handler.login)
	router.RegisterHandler("POST /v1/logout", handler.logout)
	router.RegisterHandler("GET /v1/whoami", handler.whoAmI)
}

func (handler *HttpHandler) login(w http.ResponseWriter, r *http.Request) error {
	var request loginRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		return xhttp.NewError("request does not satisfy required schema", http.StatusBadRequest)
	}

	session, err := handler.Service.CreateOrGetSessionForCredentials(r.Context(), request.Name, request.Secret)
	if err != nil {
		return err
	}

	sessionCookie := http.Cookie{
		Name:     SessionCookie,
		Value:    string(session.Id),
		Domain:   "." + handler.Domain,
		MaxAge:   handler.SessionTime,
		HttpOnly: true,
		Secure:   true,
	}
	http.SetCookie(w, &sessionCookie)
	return xhttp.WriteResponse(w, http.StatusOK, "")
}

func (handler *HttpHandler) logout(w http.ResponseWriter, r *http.Request) error {
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

func (handler *HttpHandler) whoAmI(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	identityId, ok := ctx.Value(xhttp.UserId).(string)
	if !ok {
		slog.ErrorContext(ctx, "error while parsing identity ID from context")
		return xhttp.NewError("internal server error", http.StatusInternalServerError)
	}

	identity, err := handler.Service.IdentityStore.FindById(ctx, identity.Id(identityId))
	if err != nil {
		return err
	}

	response := whoAmIResponse{
		Id:   string(identity.Id),
		Name: identity.Name,
	}

	return xhttp.WriteResponse(w, http.StatusOK, response)
}
