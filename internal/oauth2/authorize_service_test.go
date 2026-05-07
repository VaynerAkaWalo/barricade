package oauth2

import (
	"context"
	"testing"
	"time"

	"barricade/internal/authentication"
	"barricade/internal/identity"
	"barricade/internal/itest"
	"barricade/internal/keys"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
)

const (
	TEST_NAME   = "first name"
	TEST_SECRET = "changeIt"
)

type oauth2Module struct {
	authorizeService      *AuthorizeService
	authorizeHandler      *HttpHandler
	sessionService        authentication.SessionService
	identityService       identity.Service
	clientService         ClientService
	keyService            *keys.Service
	identityRepository    identity.Repository
	sessionRepository     authentication.SessionRepository
	clientRepository      ClientRepository
	authCodeRepository    AuthorizationCodeRepository
	tokenService          *TokenService
}

func setupOAuth2Module(t *testing.T) *oauth2Module {
	sessionTable := dynamodb.CreateTableInput{
		TableName: aws.String("test_session_table"),
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
			{
				AttributeName: aws.String("secondary-lookup"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("secondary-lookup-sk"),
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
						AttributeName: aws.String("secondary-lookup-sk"),
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
			{
				AttributeName: aws.String("secondary-lookup"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("secondary-lookup-sk"),
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
						AttributeName: aws.String("secondary-lookup-sk"),
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

	operationalTable := dynamodb.CreateTableInput{
		TableName: aws.String("test_operational_table"),
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
			{
				AttributeName: aws.String("secondary-lookup"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("secondary-lookup-sk"),
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
						AttributeName: aws.String("secondary-lookup-sk"),
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

	client := itest.SetupDynamo(t, sessionTable, identityTable, entitiesTable, operationalTable)

	identityStore := &identity.DynamoDBIdentityRepository{
		Client:               client,
		Table:                aws.String("test_identity_table"),
		SecondaryLookupIndex: aws.String("secondary-lookup-index"),
	}

	sessionStore := &authentication.DynamoDBSessionRepository{
		Client:               client,
		Table:                aws.String("test_session_table"),
		SecondaryLookupIndex: aws.String("secondary-lookup-index"),
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

	clientRepository := &DynamoDBClientRepository{
		Client: client,
		Table:  aws.String("test_entities_table"),
	}

	authCodeRepository := &DynamoDBAuthorizationCodeRepository{
		Client:               client,
		Table:                aws.String("test_operational_table"),
		SecondaryLookupIndex: aws.String("secondary-lookup-index"),
	}

	clientService := ClientService{Repo: clientRepository}

	keyRepo := keys.NewInMemoryRepository()
	keyService := keys.NewService(keyRepo)

	_, err := keyService.CreateKey(context.Background(), keys.RS256)
	assert.NoError(t, err)

	authorizeService := &AuthorizeService{
		ClientStore: clientRepository,
		CodeStore:   authCodeRepository,
		CodeExpiry:  5,
	}

	authService := &authentication.Service{
		IdentityStore: identityStore,
		SessionStore:  sessionStore,
	}

	authorizeHandler := &HttpHandler{
		Service:     authorizeService,
		AuthService: authService,
		LoginURL:    "https://auth.test.com/login",
	}

	tokenService := &TokenService{
		IdentityStore: identityStore,
		ClientStore:   clientRepository,
		CodeStore:     authCodeRepository,
		KeyService:    keyService,
		Issuer:        "https://test.issuer.com",
		TokenExpiry:   5,
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
		authCodeRepository: authCodeRepository,
		tokenService:       tokenService,
	}
}

func TestAuthorizeServiceGenerateCodeHappyPath(t *testing.T) {
	module := setupOAuth2Module(t)

	ident, err := module.identityService.Register(context.Background(), TEST_NAME, TEST_SECRET)
	assert.NoError(t, err)

	clientResult, err := module.clientService.Register(context.Background(), RegisterClientParams{
		OwnerId:     string(ident.Id),
		Name:        "test-app",
		Domain:      "example.com",
		RedirectURI: "https://example.com/callback",
	})
	assert.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	code, err := module.authorizeService.GenerateCode(ctx, ident.Id, AuthorizationParams{
		ClientId:    string(clientResult.Client.Id),
		RedirectURI: "https://example.com/callback",
		Scope:       "openid",
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, code)

	stored, err := module.authCodeRepository.FindByCode(ctx, code)
	assert.NoError(t, err)
	assert.Equal(t, string(clientResult.Client.Id), stored.ClientId)
	assert.Equal(t, string(ident.Id), stored.IdentityId)
	assert.Equal(t, "https://example.com/callback", stored.RedirectURI)
	assert.Equal(t, "openid", stored.Scope)
}

func TestAuthorizeServiceGenerateCodeWithPKCE(t *testing.T) {
	module := setupOAuth2Module(t)

	ident, err := module.identityService.Register(context.Background(), TEST_NAME, TEST_SECRET)
	assert.NoError(t, err)

	clientResult, err := module.clientService.Register(context.Background(), RegisterClientParams{
		OwnerId:     string(ident.Id),
		Name:        "test-app",
		Domain:      "example.com",
		RedirectURI: "https://example.com/callback",
	})
	assert.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	code, err := module.authorizeService.GenerateCode(ctx, ident.Id, AuthorizationParams{
		ClientId:            string(clientResult.Client.Id),
		RedirectURI:         "https://example.com/callback",
		Scope:               "openid",
		CodeChallenge:       "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
		CodeChallengeMethod: "S256",
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, code)

	stored, err := module.authCodeRepository.FindByCode(ctx, code)
	assert.NoError(t, err)
	assert.Equal(t, "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM", stored.CodeChallenge)
	assert.Equal(t, "S256", stored.CodeChallengeMethod)
}

func TestAuthorizeServiceGenerateCodeWithPKCEDefaultsMethod(t *testing.T) {
	module := setupOAuth2Module(t)

	ident, err := module.identityService.Register(context.Background(), TEST_NAME, TEST_SECRET)
	assert.NoError(t, err)

	clientResult, err := module.clientService.Register(context.Background(), RegisterClientParams{
		OwnerId:     string(ident.Id),
		Name:        "test-app",
		Domain:      "example.com",
		RedirectURI: "https://example.com/callback",
	})
	assert.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	code, err := module.authorizeService.GenerateCode(ctx, ident.Id, AuthorizationParams{
		ClientId:      string(clientResult.Client.Id),
		RedirectURI:   "https://example.com/callback",
		Scope:         "openid",
		CodeChallenge: "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, code)

	stored, err := module.authCodeRepository.FindByCode(ctx, code)
	assert.NoError(t, err)
	assert.Equal(t, "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM", stored.CodeChallenge)
	assert.Equal(t, "S256", stored.CodeChallengeMethod)
}

func TestValidateClientRedirectUnregisteredClient(t *testing.T) {
	module := setupOAuth2Module(t)

	params := AuthorizationParams{
		ClientId:    "nonexistent-client-id",
		RedirectURI: "https://example.com/callback",
	}

	_, _, err := module.authorizeService.ValidateClientRedirect(context.Background(), params)
	assert.ErrorIs(t, err, ErrUnauthorizedClient)
}

func TestValidateClientRedirectRedirectURIDomainMismatch(t *testing.T) {
	module := setupOAuth2Module(t)

	_, err := module.identityService.Register(context.Background(), TEST_NAME, TEST_SECRET)
	assert.NoError(t, err)

	clientResult, err := module.clientService.Register(context.Background(), RegisterClientParams{
		OwnerId:     TEST_CLIENT_OWNER_ID,
		Name:        "test-app",
		Domain:      "example.com",
		RedirectURI: "https://example.com/callback",
	})
	assert.NoError(t, err)

	params := AuthorizationParams{
		ClientId:    string(clientResult.Client.Id),
		RedirectURI: "https://evil.com/callback",
	}

	_, _, err = module.authorizeService.ValidateClientRedirect(context.Background(), params)
	assert.ErrorIs(t, err, ErrRedirectURIMismatch)
}

func TestValidateClientRedirectUsesRegisteredURIAsFallback(t *testing.T) {
	module := setupOAuth2Module(t)

	_, err := module.identityService.Register(context.Background(), TEST_NAME, TEST_SECRET)
	assert.NoError(t, err)

	clientResult, err := module.clientService.Register(context.Background(), RegisterClientParams{
		OwnerId:     TEST_CLIENT_OWNER_ID,
		Name:        "test-app",
		Domain:      "example.com",
		RedirectURI: "https://example.com/callback",
	})
	assert.NoError(t, err)

	params := AuthorizationParams{
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

	clientResult, err := module.clientService.Register(context.Background(), RegisterClientParams{
		OwnerId:     TEST_CLIENT_OWNER_ID,
		Name:        "test-app",
		Domain:      "example.com",
		RedirectURI: "https://example.com/callback",
	})
	assert.NoError(t, err)

	params := AuthorizationParams{
		ClientId:    string(clientResult.Client.Id),
		RedirectURI: "https://sub.example.com/callback",
	}

	client, redirectURI, err := module.authorizeService.ValidateClientRedirect(context.Background(), params)
	assert.NoError(t, err)
	assert.Equal(t, clientResult.Client.Id, client.Id)
	assert.Equal(t, "https://sub.example.com/callback", redirectURI)
}
