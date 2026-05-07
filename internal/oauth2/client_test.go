package oauth2

import (
	"context"
	"testing"

	"barricade/internal/itest"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

const (
	TEST_CLIENT_OWNER_ID     = "01234567-89ab-cdef-0123-456789abcdef"
	TEST_CLIENT_NAME         = "test-app"
	TEST_CLIENT_DOMAIN       = "example.com"
	TEST_CLIENT_REDIRECT_URI = "https://example.com/callback"
)

type clientModule struct {
	service    ClientService
	repository ClientRepository
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
			{
				AttributeName: aws.String("sharded-type"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{
			{
				IndexName: aws.String("sharded-type-index"),
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: aws.String("sharded-type"),
						KeyType:       types.KeyTypeHash,
					},
					{
						AttributeName: aws.String("id"),
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

	ddbClient := itest.SetupDynamo(t, entitiesTable)

	clientRepo := &DynamoDBClientRepository{
		Client:           ddbClient,
		Table:            aws.String("test_entities_table"),
		ShardedTypeIndex: aws.String("sharded-type-index"),
	}

	return &clientModule{
		service:    ClientService{Repo: clientRepo},
		repository: clientRepo,
	}
}

func TestClientRegisterEmptyOwnerId(t *testing.T) {
	_, _, err := NewClient("", TEST_CLIENT_NAME, TEST_CLIENT_DOMAIN, TEST_CLIENT_REDIRECT_URI, ClientTypeConfidential)
	assert.ErrorIs(t, err, ErrClientEmptyOwnerId)
}

func TestClientRegisterInputValidation(t *testing.T) {
	_, _, err := NewClient(TEST_CLIENT_OWNER_ID, "", TEST_CLIENT_DOMAIN, TEST_CLIENT_REDIRECT_URI, ClientTypeConfidential)
	assert.ErrorIs(t, err, ErrClientEmptyName)

	_, _, err = NewClient(TEST_CLIENT_OWNER_ID, TEST_CLIENT_NAME, "", TEST_CLIENT_REDIRECT_URI, ClientTypeConfidential)
	assert.ErrorIs(t, err, ErrClientEmptyDomain)

	_, _, err = NewClient(TEST_CLIENT_OWNER_ID, TEST_CLIENT_NAME, TEST_CLIENT_DOMAIN, "", ClientTypeConfidential)
	assert.ErrorIs(t, err, ErrClientEmptyRedirectURI)
}

func TestClientRegisterInvalidRedirectURI(t *testing.T) {
	_, _, err := NewClient(TEST_CLIENT_OWNER_ID, TEST_CLIENT_NAME, TEST_CLIENT_DOMAIN, "not-a-url", ClientTypeConfidential)
	assert.ErrorIs(t, err, ErrClientInvalidRedirectURI)
}

func TestClientRegisterRedirectURIDomainMismatch(t *testing.T) {
	_, _, err := NewClient(TEST_CLIENT_OWNER_ID, TEST_CLIENT_NAME, TEST_CLIENT_DOMAIN, "https://other.com/callback", ClientTypeConfidential)
	assert.ErrorIs(t, err, ErrClientRedirectURIDomainMismatch)
}

func TestClientRegisterInvalidClientType(t *testing.T) {
	_, _, err := NewClient(TEST_CLIENT_OWNER_ID, TEST_CLIENT_NAME, TEST_CLIENT_DOMAIN, TEST_CLIENT_REDIRECT_URI, "invalid")
	assert.ErrorIs(t, err, ErrInvalidClientType)

	_, _, err = NewClient(TEST_CLIENT_OWNER_ID, TEST_CLIENT_NAME, TEST_CLIENT_DOMAIN, TEST_CLIENT_REDIRECT_URI, ClientType(""))
	assert.ErrorIs(t, err, ErrInvalidClientType)
}

func TestClientRegisterPublicClientNoSecret(t *testing.T) {
	module := setupClientModule(t)

	result, err := module.service.Register(context.Background(), RegisterClientParams{
		OwnerId:     TEST_CLIENT_OWNER_ID,
		Name:        TEST_CLIENT_NAME,
		Domain:      TEST_CLIENT_DOMAIN,
		RedirectURI: TEST_CLIENT_REDIRECT_URI,
		ClientType:  ClientTypePublic,
	})
	assert.NoError(t, err)
	assert.Empty(t, result.ClientSecret)
	assert.Nil(t, result.Client.SecretHash)
	assert.Equal(t, ClientTypePublic, result.Client.Type)

	stored, err := module.repository.FindById(context.Background(), result.Client.Id)
	assert.NoError(t, err)
	assert.Equal(t, result.Client, stored)
}

func TestClientRegisterSubdomainAllowed(t *testing.T) {
	module := setupClientModule(t)

	result, err := module.service.Register(context.Background(), RegisterClientParams{
		OwnerId:     TEST_CLIENT_OWNER_ID,
		Name:        TEST_CLIENT_NAME,
		Domain:      TEST_CLIENT_DOMAIN,
		RedirectURI: "https://sub.example.com/callback",
		ClientType:  ClientTypeConfidential,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, result.Client.Id)
}

func TestClientRegisterHappyPath(t *testing.T) {
	module := setupClientModule(t)

	result, err := module.service.Register(context.Background(), RegisterClientParams{
		OwnerId:     TEST_CLIENT_OWNER_ID,
		Name:        TEST_CLIENT_NAME,
		Domain:      TEST_CLIENT_DOMAIN,
		RedirectURI: TEST_CLIENT_REDIRECT_URI,
		ClientType:  ClientTypeConfidential,
	})
	assert.NoError(t, err)

	assert.Equal(t, TEST_CLIENT_NAME, result.Client.Name)
	assert.Equal(t, TEST_CLIENT_OWNER_ID, result.Client.OwnerId)
	assert.Equal(t, TEST_CLIENT_DOMAIN, result.Client.Domain)
	assert.Equal(t, TEST_CLIENT_REDIRECT_URI, result.Client.RedirectURI)
	assert.Equal(t, ClientTypeConfidential, result.Client.Type)
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

	_, err := module.repository.FindById(context.Background(), ClientId("nonexistent"))
	assert.ErrorIs(t, err, ErrClientNotFound)
}

func TestClientFindAllHappyPath(t *testing.T) {
	module := setupClientModule(t)

	_, err := module.service.Register(context.Background(), RegisterClientParams{
		OwnerId:     TEST_CLIENT_OWNER_ID,
		Name:        "app-1",
		Domain:      TEST_CLIENT_DOMAIN,
		RedirectURI: TEST_CLIENT_REDIRECT_URI,
		ClientType:  ClientTypeConfidential,
	})
	assert.NoError(t, err)

	_, err = module.service.Register(context.Background(), RegisterClientParams{
		OwnerId:     TEST_CLIENT_OWNER_ID,
		Name:        "app-2",
		Domain:      TEST_CLIENT_DOMAIN,
		RedirectURI: TEST_CLIENT_REDIRECT_URI,
		ClientType:  ClientTypeConfidential,
	})
	assert.NoError(t, err)

	_, err = module.service.Register(context.Background(), RegisterClientParams{
		OwnerId:     "other-owner",
		Name:        "app-3",
		Domain:      TEST_CLIENT_DOMAIN,
		RedirectURI: TEST_CLIENT_REDIRECT_URI,
		ClientType:  ClientTypeConfidential,
	})
	assert.NoError(t, err)

	clients, err := module.service.FindAll(context.Background())
	assert.NoError(t, err)
	assert.Len(t, clients, 3)
}
