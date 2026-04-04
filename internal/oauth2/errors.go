package oauth2

import "errors"

var (
	ErrInvalidRequest          = errors.New("invalid request")
	ErrUnsupportedResponseType = errors.New("unsupported response type")
	ErrInvalidScope            = errors.New("invalid scope")
	ErrInvalidRedirectURI      = errors.New("invalid redirect uri")
	ErrServerError             = errors.New("server error")
	ErrLoginRequired           = errors.New("login required")
)
