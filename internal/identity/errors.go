package identity

import "errors"

var (
	ErrIdentityNotFound      = errors.New("identity not found")
	ErrDuplicateIdentityName = errors.New("identity with this name already exists")
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrEmptyName             = errors.New("name cannot be empty")
	ErrEmptySecret           = errors.New("secret cannot be empty")
)
