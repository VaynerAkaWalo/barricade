package oidc

import (
	"net/http"

	"github.com/VaynerAkaWalo/go-toolkit/xhttp"
)

type DiscoveryResponse struct {
	Issuer                           string   `json:"issuer"`
	AuthorizationEndpoint            string   `json:"authorization_endpoint"`
	TokenEndpoint                    string   `json:"token_endpoint"`
	UserinfoEndpoint                 string   `json:"userinfo_endpoint"`
	JWKSUri                          string   `json:"jwks_uri"`
	ScopesSupported                  []string `json:"scopes_supported"`
	ResponseTypesSupported           []string `json:"response_types_supported"`
	ResponseModesSupported           []string `json:"response_modes_supported"`
	SubjectTypesSupported            []string `json:"subject_types_supported"`
	IDTokenSigningAlgValuesSupported []string `json:"id_token_signing_alg_values_supported"`
	ClaimsSupported                  []string `json:"claims_supported"`
	CodeChallengeMethodsSupported    []string `json:"code_challenge_methods_supported"`
}

type DiscoveryHandler struct {
	Issuer string
}

func (h *DiscoveryHandler) RegisterRoutes(router *xhttp.Router) {
	router.RegisterHandler("GET /.well-known/openid-configuration", h.GetDiscovery)
}

func (h *DiscoveryHandler) GetDiscovery(w http.ResponseWriter, r *http.Request) error {
	response := DiscoveryResponse{
		Issuer:                h.Issuer,
		AuthorizationEndpoint: h.Issuer + "/v1/oauth2/authorize",
		TokenEndpoint:         h.Issuer + "/v1/oauth2/token",
		UserinfoEndpoint:      h.Issuer + "/v1/oauth2/userinfo",
		JWKSUri:               h.Issuer + "/.well-known/jwks.json",
		ScopesSupported:       []string{"openid", "profile"},
		ResponseTypesSupported: []string{"code"},
		ResponseModesSupported: []string{"query", "fragment"},
		SubjectTypesSupported:  []string{"public"},
		IDTokenSigningAlgValuesSupported: []string{"RS256"},
		ClaimsSupported: []string{
			"sub", "name", "iss", "aud", "exp", "iat",
			"auth_time", "acr", "amr", "nonce",
		},
		CodeChallengeMethodsSupported: []string{"S256"},
	}

	return xhttp.WriteResponse(w, http.StatusOK, response)
}
