package oauth2

import (
	"barricade/internal/identity"
	"barricade/internal/keys"
	"fmt"
	"net/http"
	"strings"

	"github.com/VaynerAkaWalo/go-toolkit/xhttp"
	"github.com/golang-jwt/jwt/v5"
)

type userinfoResponse struct {
	Sub  string `json:"sub"`
	Name string `json:"name,omitempty"`
}

type UserinfoHandler struct {
	KeyService    *keys.Service
	IdentityStore IdentityRepository
	Issuer        string
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

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("missing kid")
		}

		key, err := h.KeyService.GetKey(ctx, keys.KeyId(kid))
		if err != nil {
			return nil, fmt.Errorf("key not found: %w", err)
		}

		return key.RSAPublicKey()
	})
	if err != nil {
		w.Header().Set("WWW-Authenticate", `Bearer error="invalid_token", error_description="access token is invalid or expired"`)
		return xhttp.NewError("invalid_token", http.StatusUnauthorized)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		w.Header().Set("WWW-Authenticate", `Bearer error="invalid_token", error_description="access token is invalid"`)
		return xhttp.NewError("invalid_token", http.StatusUnauthorized)
	}

	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		w.Header().Set("WWW-Authenticate", `Bearer error="invalid_token", error_description="access token missing subject"`)
		return xhttp.NewError("invalid_token", http.StatusUnauthorized)
	}

	issuer, _ := claims["iss"].(string)
	if issuer != h.Issuer {
		w.Header().Set("WWW-Authenticate", `Bearer error="invalid_token", error_description="access token issuer mismatch"`)
		return xhttp.NewError("invalid_token", http.StatusUnauthorized)
	}

	scope, _ := claims["scope"].(string)

	if !strings.Contains(scope, "openid") {
		w.Header().Set("WWW-Authenticate", `Bearer error="insufficient_scope", scope="openid"`)
		return xhttp.NewError("insufficient_scope", http.StatusForbidden)
	}

	ident, err := h.IdentityStore.FindById(ctx, identity.Id(sub))
	if err != nil {
		w.Header().Set("WWW-Authenticate", `Bearer error="invalid_token", error_description="identity not found"`)
		return xhttp.NewError("invalid_token", http.StatusUnauthorized)
	}

	resp := userinfoResponse{
		Sub: string(ident.Id),
	}

	if strings.Contains(scope, "profile") {
		resp.Name = ident.Name
	}

	return xhttp.WriteResponse(w, http.StatusOK, resp)
}
