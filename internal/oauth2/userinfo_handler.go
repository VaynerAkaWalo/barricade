package oauth2

import (
	"errors"
	"net/http"
	"strings"

	"github.com/VaynerAkaWalo/go-toolkit/xhttp"
)

type userinfoResponse struct {
	Sub  string `json:"sub"`
	Name string `json:"name,omitempty"`
}

type UserinfoHandler struct {
	Service *UserinfoService
}

func (h *UserinfoHandler) RegisterRoutes(router *xhttp.Router) {
	router.RegisterHandler("GET /v1/oauth2/userinfo", h.Userinfo)
}

func (h *UserinfoHandler) Userinfo(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		w.Header().Set("WWW-Authenticate", `Bearer error="invalid_token", error_description="missing or malformed authorization header"`)
		return xhttp.NewError("invalid_token", http.StatusUnauthorized)
	}

	accessToken := strings.TrimPrefix(authHeader, "Bearer ")

	result, err := h.Service.GetUserinfo(ctx, accessToken)
	if err != nil {
		code, status := mapUserinfoError(err)
		if desc := errorDescription(err); desc != "" {
			w.Header().Set("WWW-Authenticate", `Bearer error="`+code+`", error_description="`+desc+`"`)
		} else {
			w.Header().Set("WWW-Authenticate", userinfoWWWAuthenticate(code))
		}
		return xhttp.NewError(code, status)
	}

	resp := userinfoResponse{
		Sub:  result.Sub,
		Name: result.Name,
	}

	return xhttp.WriteResponse(w, http.StatusOK, resp)
}

func errorDescription(err error) string {
	switch {
	case errors.Is(err, ErrInsufficientScope):
		return ""
	default:
		return "access token is invalid"
	}
}
