package oidc

import (
	"barricade/internal/keys"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"log/slog"
	"math/big"
)

func keyToJWK(key *keys.Key) (*JWK, error) {
	block, _ := pem.Decode(key.PublicKey)
	if block == nil {
		return nil, fmt.Errorf("failed to parse public key PEM")
	}

	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	var kty, alg string
	switch key.Algorithm {
	case keys.RS256:
		kty = "RSA"
		alg = "RS256"
		_, ok := pubKey.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("public key is not RSA")
		}
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", key.Algorithm)
	}

	rsaPubKey := pubKey.(*rsa.PublicKey)
	n := base64.RawURLEncoding.EncodeToString(rsaPubKey.N.Bytes())
	e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(rsaPubKey.E)).Bytes())

	return &JWK{
		Kty: kty,
		Kid: string(key.Id),
		Use: "sig",
		Alg: alg,
		N:   n,
		E:   e,
	}, nil
}

func filterRSAKeys(allKeys []*keys.Key) []*keys.Key {
	var result []*keys.Key
	for _, key := range allKeys {
		if key.Algorithm == keys.RS256 {
			result = append(result, key)
		}
	}
	return result
}

func keysToJWKS(allKeys []*keys.Key) *JWKSResponse {
	rsaKeys := filterRSAKeys(allKeys)
	jwks := make([]JWK, 0, len(rsaKeys))

	for _, key := range rsaKeys {
		jwk, err := keyToJWK(key)
		if err != nil {
			slog.Error("failed to convert key to JWK", "keyId", key.Id, "error", err)
			continue
		}
		jwks = append(jwks, *jwk)
	}

	return &JWKSResponse{Keys: jwks}
}
