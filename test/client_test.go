package test

import (
	"barricade/internal/oauth2"
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

const (
	TEST_CLIENT_OWNER_ID    = "01234567-89ab-cdef-0123-456789abcdef"
	TEST_CLIENT_NAME         = "test-app"
	TEST_CLIENT_DOMAIN       = "example.com"
	TEST_CLIENT_REDIRECT_URI = "https://example.com/callback"
)

type clientModule struct {
	service    oauth2.ClientService
	repository oauth2.ClientRepository
}

func setupClientModule(t *testing.T) *clientModule {
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

	ddbClient := setupDynamo(t, entitiesTable)

	clientRepo := &oauth2.DynamoDBClientRepository{
		Client: ddbClient,
		Table:  aws.String("test_entities_table"),
	}

	return &clientModule{
		service:    oauth2.ClientService{Repo: clientRepo},
		repository: clientRepo,
	}
}

func TestClientRegisterEmptyOwnerId(t *testing.T) {
	module := setupClientModule(t)

	_, err := module.service.Register(context.Background(), oauth2.RegisterClientParams{
		OwnerId:     "",
		Name:        TEST_CLIENT_NAME,
		Domain:      TEST_CLIENT_DOMAIN,
		RedirectURI: TEST_CLIENT_REDIRECT_URI,
	})
	assert.ErrorIs(t, err, oauth2.ErrClientEmptyOwnerId)
}

func TestClientRegisterInputValidation(t *testing.T) {
	module := setupClientModule(t)

	_, err := module.service.Register(context.Background(), oauth2.RegisterClientParams{
		OwnerId:     TEST_CLIENT_OWNER_ID,
		Name:        "",
		Domain:      TEST_CLIENT_DOMAIN,
		RedirectURI: TEST_CLIENT_REDIRECT_URI,
	})
	assert.ErrorIs(t, err, oauth2.ErrClientEmptyName)

	_, err = module.service.Register(context.Background(), oauth2.RegisterClientParams{
		OwnerId:     TEST_CLIENT_OWNER_ID,
		Name:        TEST_CLIENT_NAME,
		Domain:      "",
		RedirectURI: TEST_CLIENT_REDIRECT_URI,
	})
	assert.ErrorIs(t, err, oauth2.ErrClientEmptyDomain)

	_, err = module.service.Register(context.Background(), oauth2.RegisterClientParams{
		OwnerId:     TEST_CLIENT_OWNER_ID,
		Name:        TEST_CLIENT_NAME,
		Domain:      TEST_CLIENT_DOMAIN,
		RedirectURI: "",
	})
	assert.ErrorIs(t, err, oauth2.ErrClientEmptyRedirectURI)
}

func TestClientRegisterInvalidRedirectURI(t *testing.T) {
	module := setupClientModule(t)

	_, err := module.service.Register(context.Background(), oauth2.RegisterClientParams{
		OwnerId:     TEST_CLIENT_OWNER_ID,
		Name:        TEST_CLIENT_NAME,
		Domain:      TEST_CLIENT_DOMAIN,
		RedirectURI: "not-a-url",
	})
	assert.ErrorIs(t, err, oauth2.ErrClientInvalidRedirectURI)
}

func TestClientRegisterRedirectURIDomainMismatch(t *testing.T) {
	module := setupClientModule(t)

	_, err := module.service.Register(context.Background(), oauth2.RegisterClientParams{
		OwnerId:     TEST_CLIENT_OWNER_ID,
		Name:        TEST_CLIENT_NAME,
		Domain:      TEST_CLIENT_DOMAIN,
		RedirectURI: "https://other.com/callback",
	})
	assert.ErrorIs(t, err, oauth2.ErrClientRedirectURIDomainMismatch)
}

func TestClientRegisterSubdomainAllowed(t *testing.T) {
	module := setupClientModule(t)

	result, err := module.service.Register(context.Background(), oauth2.RegisterClientParams{
		OwnerId:     TEST_CLIENT_OWNER_ID,
		Name:        TEST_CLIENT_NAME,
		Domain:      TEST_CLIENT_DOMAIN,
		RedirectURI: "https://sub.example.com/callback",
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, result.Client.Id)
}

func TestClientRegisterHappyPath(t *testing.T) {
	module := setupClientModule(t)

	result, err := module.service.Register(context.Background(), oauth2.RegisterClientParams{
		OwnerId:     TEST_CLIENT_OWNER_ID,
		Name:        TEST_CLIENT_NAME,
		Domain:      TEST_CLIENT_DOMAIN,
		RedirectURI: TEST_CLIENT_REDIRECT_URI,
	})
	assert.NoError(t, err)

	assert.Equal(t, TEST_CLIENT_NAME, result.Client.Name)
	assert.Equal(t, TEST_CLIENT_OWNER_ID, result.Client.OwnerId)
	assert.Equal(t, TEST_CLIENT_DOMAIN, result.Client.Domain)
	assert.Equal(t, TEST_CLIENT_REDIRECT_URI, result.Client.RedirectURI)
	assert.NotEmpty(t, result.Client.Id)
	assert.NotEmpty(t, result.ClientSecret)
	assert.Nil(t, bcrypt.CompareHashAndPassword(result.Client.SecretHash, []byte(result.ClientSecret)))
	assert.NotEmpty(t, result.Client.CreatedAt)
	assert.NotEmpty(t, result.Client.UpdatedAt)

	stored, err := module.repository.FindById(context.Background(), result.Client.Id)
	assert.NoError(t, err)
	assert.Equal(t, result.Client, stored)
}

func TestClientFindByIdNotFound(t *testing.T) {
	module := setupClientModule(t)

	_, err := module.repository.FindById(context.Background(), oauth2.ClientId("nonexistent"))
	assert.ErrorIs(t, err, oauth2.ErrClientNotFound)
}
