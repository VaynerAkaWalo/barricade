package authentication

import "errors"

var (
	ErrSessionNotFound     = errors.New("session not found")
	ErrSessionExpired      = errors.New("session expired")
	ErrEmptySessionId      = errors.New("session id cannot be empty")
	ErrInvalidSessionOwner = errors.New("invalid session owner")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrEmptyName           = errors.New("name cannot be empty")
	ErrEmptySecret         = errors.New("secret cannot be empty")
)
