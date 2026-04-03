package oidc

import (
	"barricade/internal/keys"
	"net/http"

	"github.com/VaynerAkaWalo/go-toolkit/xhttp"
)

type JWK struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type JWKSResponse struct {
	Keys []JWK `json:"keys"`
}

type JWKSHandler struct {
	KeyService *keys.Service
}

func (h *JWKSHandler) RegisterRoutes(router *xhttp.Router) {
	router.RegisterHandler("GET /.well-known/jwks.json", h.GetJWKS)
}

func (h *JWKSHandler) GetJWKS(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	allKeys, err := h.KeyService.ListAllKeys(ctx)
	if err != nil {
		return xhttp.NewError("unable to retrieve keys", http.StatusInternalServerError)
	}

	jwks := keysToJWKS(allKeys)

	return xhttp.WriteResponse(w, http.StatusOK, jwks)
}
