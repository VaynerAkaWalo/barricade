package oauth2

import "errors"

var (
	ErrInvalidRequest                  = errors.New("invalid request")
	ErrUnsupportedResponseType         = errors.New("unsupported response type")
	ErrInvalidScope                    = errors.New("invalid scope")
	ErrInvalidRedirectURI              = errors.New("invalid redirect uri")
	ErrServerError                     = errors.New("server error")
	ErrLoginRequired                   = errors.New("login required")
	ErrClientNotFound                  = errors.New("client not found")
	ErrClientEmptyOwnerId              = errors.New("client owner id cannot be empty")
	ErrClientEmptyName                 = errors.New("client name cannot be empty")
	ErrClientEmptyDomain               = errors.New("client domain cannot be empty")
	ErrClientEmptyRedirectURI          = errors.New("client redirect URI cannot be empty")
	ErrClientInvalidRedirectURI        = errors.New("client redirect URI is not a valid URL")
	ErrClientRedirectURIDomainMismatch = errors.New("client redirect URI domain does not match client domain")
)
