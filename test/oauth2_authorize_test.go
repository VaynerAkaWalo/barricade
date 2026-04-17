package test

import (
	"barricade/internal/authentication"
	"barricade/internal/identity"
	"barricade/internal/keys"
	"barricade/internal/oauth2"
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
)

type oauth2Module struct {
	authorizeService   *oauth2.AuthorizeService
	authorizeHandler   *oauth2.HttpHandler
	sessionService     authentication.SessionService
	identityService    identity.Service
	clientService      oauth2.ClientService
	keyService         *keys.Service
	identityRepository identity.Repository
	sessionRepository  authentication.SessionRepository
	clientRepository   oauth2.ClientRepository
}

func setupOAuth2Module(t *testing.T) *oauth2Module {
	sessionTable := dynamodb.CreateTableInput{
		TableName: aws.String("test_session_table"),
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("id"),
				KeyType:       types.KeyTypeHash,
			},
		},
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("id"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("secondary-lookup"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("resource-type"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{
			{
				IndexName: aws.String("secondary-lookup-index"),
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: aws.String("secondary-lookup"),
						KeyType:       types.KeyTypeHash,
					},
					{
						AttributeName: aws.String("resource-type"),
						KeyType:       types.KeyTypeRange,
					},
				},
				Projection: &types.Projection{
					ProjectionType: types.ProjectionTypeAll,
				},
			},
		},
		BillingMode: types.BillingModePayPerRequest,
	}

	identityTable := dynamodb.CreateTableInput{
		TableName: aws.String("test_identity_table"),
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("id"),
				KeyType:       types.KeyTypeHash,
			},
		},
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("id"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("name"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{
			{
				IndexName: aws.String("name-index"),
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: aws.String("name"),
						KeyType:       types.KeyTypeHash,
					},
				},
				Projection: &types.Projection{
					NonKeyAttributes: []string{"id", "secret"},
					ProjectionType:   types.ProjectionTypeInclude,
				},
			},
		},
		BillingMode: types.BillingModePayPerRequest,
	}

	entitiesTable := dynamodb.CreateTableInput{
		TableName: aws.String("test_entities_table"),
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("id"),
				KeyType:       types.KeyTypeHash,
			},
			{
				AttributeName: aws.String("type"),
				KeyType:       types.KeyTypeRange,
			},
		},
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("id"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("type"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		BillingMode: types.BillingModePayPerRequest,
	}

	client := setupDynamo(t, sessionTable, identityTable, entitiesTable)

	identityStore := &identity.DynamoDBIdentityRepository{
		Client:    client,
		Table:     aws.String("test_identity_table"),
		NameIndex: aws.String("name-index"),
	}

	sessionStore := &authentication.DynamoDBSessionRepository{
		Client:    client,
		Table:     aws.String("test_session_table"),
		NameIndex: aws.String("secondary-lookup-index"),
	}

	sessionService := authentication.SessionService{
		SessionStore:  sessionStore,
		IdentityStore: identityStore,
	}

	identityService := identity.Service{
		Repo: &identity.DynamoDBIdentityRepository{
			Client: client,
			Table:  aws.String("test_identity_table"),
		},
	}

	clientRepository := &oauth2.DynamoDBClientRepository{
		Client: client,
		Table:  aws.String("test_entities_table"),
	}

	clientService := oauth2.ClientService{Repo: clientRepository}

	keyRepo := keys.NewInMemoryRepository()
	keyService := keys.NewService(keyRepo)

	_, err := keyService.CreateKey(context.Background(), keys.RS256)
	assert.NoError(t, err)

	authorizeService := &oauth2.AuthorizeService{
		IdentityStore: identityStore,
		ClientStore:   clientRepository,
		KeyService:    keyService,
		Issuer:        "https://test.issuer.com",
		TokenExpiry:   5,
	}

	authService := &authentication.Service{
		IdentityStore: identityStore,
		SessionStore:  sessionStore,
	}

	authorizeHandler := &oauth2.HttpHandler{
		Service:     authorizeService,
		AuthService: authService,
		LoginURL:    "https://auth.test.com/login",
	}

	return &oauth2Module{
		authorizeService:   authorizeService,
		authorizeHandler:   authorizeHandler,
		sessionService:     sessionService,
		identityService:    identityService,
		clientService:      clientService,
		keyService:         keyService,
		identityRepository: identityStore,
		sessionRepository:  sessionStore,
		clientRepository:   clientRepository,
	}
}

func TestAuthorizeServiceValidateMissingResponseType(t *testing.T) {
	module := setupOAuth2Module(t)

	params := oauth2.AuthorizationParams{
		ClientId: "test-client",
		Scope:    "openid",
	}

	err := module.authorizeService.Validate(params)
	assert.ErrorIs(t, err, oauth2.ErrInvalidRequest)
}

func TestAuthorizeServiceValidateUnsupportedResponseType(t *testing.T) {
	module := setupOAuth2Module(t)

	params := oauth2.AuthorizationParams{
		ResponseType: "code",
		ClientId:     "test-client",
		Scope:        "openid",
	}

	err := module.authorizeService.Validate(params)
	assert.ErrorIs(t, err, oauth2.ErrUnsupportedResponseType)
}

func TestAuthorizeServiceValidateMissingOpenIDScope(t *testing.T) {
	module := setupOAuth2Module(t)

	params := oauth2.AuthorizationParams{
		ResponseType: "id_token",
		ClientId:     "test-client",
		Scope:        "profile email",
	}

	err := module.authorizeService.Validate(params)
	assert.ErrorIs(t, err, oauth2.ErrInvalidScope)
}

func TestAuthorizeServiceValidateHappyPath(t *testing.T) {
	module := setupOAuth2Module(t)

	params := oauth2.AuthorizationParams{
		ResponseType: "id_token",
		ClientId:     "test-client",
		Scope:        "openid profile",
	}

	err := module.authorizeService.Validate(params)
	assert.NoError(t, err)
}

func TestAuthorizeServiceAuthorizeHappyPath(t *testing.T) {
	module := setupOAuth2Module(t)

	ident, err := module.identityService.Register(context.Background(), TEST_NAME, TEST_SECRET)
	assert.NoError(t, err)

	clientResult, err := module.clientService.Register(context.Background(), oauth2.RegisterClientParams{
		OwnerId:     string(ident.Id),
		Name:        "test-app",
		Domain:      "example.com",
		RedirectURI: "https://example.com/callback",
	})
	assert.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := module.authorizeService.Authorize(ctx, ident.Id, string(clientResult.Client.Id))
	assert.NoError(t, err)
	assert.NotEmpty(t, result.IDToken)
	assert.NotNil(t, result)
}

func TestValidateClientRedirectUnregisteredClient(t *testing.T) {
	module := setupOAuth2Module(t)

	params := oauth2.AuthorizationParams{
		ClientId:    "nonexistent-client-id",
		RedirectURI: "https://example.com/callback",
	}

	_, _, err := module.authorizeService.ValidateClientRedirect(context.Background(), params)
	assert.ErrorIs(t, err, oauth2.ErrUnauthorizedClient)
}

func TestValidateClientRedirectRedirectURIDomainMismatch(t *testing.T) {
	module := setupOAuth2Module(t)

	_, err := module.identityService.Register(context.Background(), TEST_NAME, TEST_SECRET)
	assert.NoError(t, err)

	clientResult, err := module.clientService.Register(context.Background(), oauth2.RegisterClientParams{
		OwnerId:     TEST_CLIENT_OWNER_ID,
		Name:        "test-app",
		Domain:      "example.com",
		RedirectURI: "https://example.com/callback",
	})
	assert.NoError(t, err)

	params := oauth2.AuthorizationParams{
		ClientId:    string(clientResult.Client.Id),
		RedirectURI: "https://evil.com/callback",
	}

	_, _, err = module.authorizeService.ValidateClientRedirect(context.Background(), params)
	assert.ErrorIs(t, err, oauth2.ErrRedirectURIMismatch)
}

func TestValidateClientRedirectUsesRegisteredURIAsFallback(t *testing.T) {
	module := setupOAuth2Module(t)

	_, err := module.identityService.Register(context.Background(), TEST_NAME, TEST_SECRET)
	assert.NoError(t, err)

	clientResult, err := module.clientService.Register(context.Background(), oauth2.RegisterClientParams{
		OwnerId:     TEST_CLIENT_OWNER_ID,
		Name:        "test-app",
		Domain:      "example.com",
		RedirectURI: "https://example.com/callback",
	})
	assert.NoError(t, err)

	params := oauth2.AuthorizationParams{
		ClientId:    string(clientResult.Client.Id),
		RedirectURI: "",
	}

	client, redirectURI, err := module.authorizeService.ValidateClientRedirect(context.Background(), params)
	assert.NoError(t, err)
	assert.Equal(t, clientResult.Client.Id, client.Id)
	assert.Equal(t, "https://example.com/callback", redirectURI)
}

func TestValidateClientRedirectSubdomainAllowed(t *testing.T) {
	module := setupOAuth2Module(t)

	_, err := module.identityService.Register(context.Background(), TEST_NAME, TEST_SECRET)
	assert.NoError(t, err)

	clientResult, err := module.clientService.Register(context.Background(), oauth2.RegisterClientParams{
		OwnerId:     TEST_CLIENT_OWNER_ID,
		Name:        "test-app",
		Domain:      "example.com",
		RedirectURI: "https://example.com/callback",
	})
	assert.NoError(t, err)

	params := oauth2.AuthorizationParams{
		ClientId:    string(clientResult.Client.Id),
		RedirectURI: "https://sub.example.com/callback",
	}

	client, redirectURI, err := module.authorizeService.ValidateClientRedirect(context.Background(), params)
	assert.NoError(t, err)
	assert.Equal(t, clientResult.Client.Id, client.Id)
	assert.Equal(t, "https://sub.example.com/callback", redirectURI)
}
