package oauth2

import "time"

type AuthorizationCode struct {
	Code        string
	ClientId    string
	IdentityId  string
	RedirectURI string
	Scope       string
	CreatedAt   int64
	ExpireAt    int64
}

func NewAuthorizationCode(clientId string, identityId string, redirectURI string, scope string, expiryMinutes int) *AuthorizationCode {
	now := time.Now().Unix()
	expire := now + int64(expiryMinutes)*60

	return &AuthorizationCode{
		ClientId:    clientId,
		IdentityId:  identityId,
		RedirectURI: redirectURI,
		Scope:       scope,
		CreatedAt:   now,
		ExpireAt:    expire,
	}
}
