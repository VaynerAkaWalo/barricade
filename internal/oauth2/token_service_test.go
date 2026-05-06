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

func setupTokenModule(t *testing.T) *oauth2Module {
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

	clientRepository := &DynamoDBClientRepository{
		Client: client,
		Table:  aws.String("test_entities_table"),
	}

	authCodeRepository := &DynamoDBAuthorizationCodeRepository{
		Client:               client,
		Table:                aws.String("test_operational_table"),
		SecondaryLookupIndex: aws.String("secondary-lookup-index"),
	}

	keyRepo := keys.NewInMemoryRepository()
	keyService := keys.NewService(keyRepo)

	_, err := keyService.CreateKey(context.Background(), keys.RS256)
	assert.NoError(t, err)

	identityService := identity.Service{
		Repo: identityStore,
	}

	clientService := ClientService{Repo: clientRepository}

	tokenService := &TokenService{
		IdentityStore: identityStore,
		ClientStore:   clientRepository,
		CodeStore:     authCodeRepository,
		KeyService:    keyService,
		Issuer:        "https://test.issuer.com",
		TokenExpiry:   5,
	}

	return &oauth2Module{
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

func TestTokenExchangeHappyPath(t *testing.T) {
	module := setupTokenModule(t)

	ident, err := module.identityService.Register(context.Background(), TEST_NAME, TEST_SECRET)
	assert.NoError(t, err)

	clientResult, err := module.clientService.Register(context.Background(), RegisterClientParams{
		OwnerId:     string(ident.Id),
		Name:        "test-app",
		Domain:      "example.com",
		RedirectURI: "https://example.com/callback",
	})
	assert.NoError(t, err)

	code := NewAuthorizationCode(string(clientResult.Client.Id), string(ident.Id), "https://example.com/callback", "openid", 5)
	code.Code = "test-auth-code-123"
	err = module.authCodeRepository.Save(context.Background(), code)
	assert.NoError(t, err)

	result, err := module.tokenService.Exchange(context.Background(), ExchangeTokenParams{
		GrantType:    "authorization_code",
		Code:         "test-auth-code-123",
		RedirectURI:  "https://example.com/callback",
		ClientId:     string(clientResult.Client.Id),
		ClientSecret: string(clientResult.ClientSecret),
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, result.IDToken)
	assert.Equal(t, "Bearer", result.TokenType)
	assert.Equal(t, 300, result.ExpiresIn)
}

func TestTokenExchangeUnsupportedGrantType(t *testing.T) {
	svc := &TokenService{}

	_, err := svc.Exchange(context.Background(), ExchangeTokenParams{
		GrantType: "client_credentials",
	})
	assert.ErrorIs(t, err, ErrUnsupportedGrantType)
}

func TestTokenExchangeInvalidClient(t *testing.T) {
	module := setupTokenModule(t)

	ident, err := module.identityService.Register(context.Background(), TEST_NAME, TEST_SECRET)
	assert.NoError(t, err)

	clientResult, err := module.clientService.Register(context.Background(), RegisterClientParams{
		OwnerId:     string(ident.Id),
		Name:        "test-app",
		Domain:      "example.com",
		RedirectURI: "https://example.com/callback",
	})
	assert.NoError(t, err)

	_, err = module.tokenService.Exchange(context.Background(), ExchangeTokenParams{
		GrantType:    "authorization_code",
		ClientId:     string(clientResult.Client.Id),
		ClientSecret: "wrong-secret",
	})
	assert.ErrorIs(t, err, ErrInvalidClient)
}

func TestTokenExchangeInvalidCode(t *testing.T) {
	module := setupTokenModule(t)

	ident, err := module.identityService.Register(context.Background(), TEST_NAME, TEST_SECRET)
	assert.NoError(t, err)

	clientResult, err := module.clientService.Register(context.Background(), RegisterClientParams{
		OwnerId:     string(ident.Id),
		Name:        "test-app",
		Domain:      "example.com",
		RedirectURI: "https://example.com/callback",
	})
	assert.NoError(t, err)

	_, err = module.tokenService.Exchange(context.Background(), ExchangeTokenParams{
		GrantType:    "authorization_code",
		Code:         "nonexistent-code",
		ClientId:     string(clientResult.Client.Id),
		ClientSecret: string(clientResult.ClientSecret),
	})
	assert.ErrorIs(t, err, ErrInvalidCode)
}

func TestTokenExchangeExpiredCode(t *testing.T) {
	module := setupTokenModule(t)

	ident, err := module.identityService.Register(context.Background(), TEST_NAME, TEST_SECRET)
	assert.NoError(t, err)

	clientResult, err := module.clientService.Register(context.Background(), RegisterClientParams{
		OwnerId:     string(ident.Id),
		Name:        "test-app",
		Domain:      "example.com",
		RedirectURI: "https://example.com/callback",
	})
	assert.NoError(t, err)

	code := NewAuthorizationCode(string(clientResult.Client.Id), string(ident.Id), "https://example.com/callback", "openid", 5)
	code.Code = "expired-code-123"
	code.ExpireAt = time.Now().Add(-1 * time.Minute).Unix()
	err = module.authCodeRepository.Save(context.Background(), code)
	assert.NoError(t, err)

	_, err = module.tokenService.Exchange(context.Background(), ExchangeTokenParams{
		GrantType:    "authorization_code",
		Code:         "expired-code-123",
		ClientId:     string(clientResult.Client.Id),
		ClientSecret: string(clientResult.ClientSecret),
	})
	assert.ErrorIs(t, err, ErrCodeExpired)
}

func TestTokenExchangeRedirectURIMismatch(t *testing.T) {
	module := setupTokenModule(t)

	ident, err := module.identityService.Register(context.Background(), TEST_NAME, TEST_SECRET)
	assert.NoError(t, err)

	clientResult, err := module.clientService.Register(context.Background(), RegisterClientParams{
		OwnerId:     string(ident.Id),
		Name:        "test-app",
		Domain:      "example.com",
		RedirectURI: "https://example.com/callback",
	})
	assert.NoError(t, err)

	code := NewAuthorizationCode(string(clientResult.Client.Id), string(ident.Id), "https://example.com/callback", "openid", 5)
	code.Code = "mismatch-code-123"
	err = module.authCodeRepository.Save(context.Background(), code)
	assert.NoError(t, err)

	_, err = module.tokenService.Exchange(context.Background(), ExchangeTokenParams{
		GrantType:    "authorization_code",
		Code:         "mismatch-code-123",
		RedirectURI:  "https://evil.com/callback",
		ClientId:     string(clientResult.Client.Id),
		ClientSecret: string(clientResult.ClientSecret),
	})
	assert.ErrorIs(t, err, ErrCodeMismatch)
}

func TestTokenExchangeCodeReplay(t *testing.T) {
	module := setupTokenModule(t)

	ident, err := module.identityService.Register(context.Background(), TEST_NAME, TEST_SECRET)
	assert.NoError(t, err)

	clientResult, err := module.clientService.Register(context.Background(), RegisterClientParams{
		OwnerId:     string(ident.Id),
		Name:        "test-app",
		Domain:      "example.com",
		RedirectURI: "https://example.com/callback",
	})
	assert.NoError(t, err)

	code := NewAuthorizationCode(string(clientResult.Client.Id), string(ident.Id), "https://example.com/callback", "openid", 5)
	code.Code = "replay-code-123"
	err = module.authCodeRepository.Save(context.Background(), code)
	assert.NoError(t, err)

	_, err = module.tokenService.Exchange(context.Background(), ExchangeTokenParams{
		GrantType:    "authorization_code",
		Code:         "replay-code-123",
		RedirectURI:  "https://example.com/callback",
		ClientId:     string(clientResult.Client.Id),
		ClientSecret: string(clientResult.ClientSecret),
	})
	assert.NoError(t, err)

	_, err = module.tokenService.Exchange(context.Background(), ExchangeTokenParams{
		GrantType:    "authorization_code",
		Code:         "replay-code-123",
		RedirectURI:  "https://example.com/callback",
		ClientId:     string(clientResult.Client.Id),
		ClientSecret: string(clientResult.ClientSecret),
	})
	assert.ErrorIs(t, err, ErrInvalidCode)
}
