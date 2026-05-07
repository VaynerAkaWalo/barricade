package oauth2

import "context"

type ClientRepository interface {
	Save(ctx context.Context, client *Client) error
	FindById(ctx context.Context, id ClientId) (*Client, error)
	FindAll(ctx context.Context) ([]*Client, error)
}

type RegisterClientParams struct {
	OwnerId     string
	Name        string
	Domain      string
	RedirectURI string
	ClientType  ClientType
}

type RegisterClientResult struct {
	Client       *Client
	ClientSecret ClientSecret
}

type ClientService struct {
	Repo ClientRepository
}

func (s *ClientService) Register(ctx context.Context, params RegisterClientParams) (*RegisterClientResult, error) {
	c, secret, err := NewClient(params.OwnerId, params.Name, params.Domain, params.RedirectURI, params.ClientType)
	if err != nil {
		return nil, err
	}

	err = s.Repo.Save(ctx, c)
	if err != nil {
		return nil, err
	}

	return &RegisterClientResult{
		Client:       c,
		ClientSecret: secret,
	}, nil
}

func (s *ClientService) FindById(ctx context.Context, id ClientId) (*Client, error) {
	return s.Repo.FindById(ctx, id)
}

func (s *ClientService) FindAll(ctx context.Context) ([]*Client, error) {
	return s.Repo.FindAll(ctx)
}
