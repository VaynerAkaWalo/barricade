package test

import (
	dynamodbadapters "barricade/internal/adapters/dynamodb"
	"barricade/internal/domain/authentication"
	"barricade/internal/domain/identity"
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

type sessionModule struct {
	sessionService              authentication.SessionService
	sessionStore                authentication.SessionRepository
	authenticationIdentityStore authentication.IdentityRepository
	identityService             identity.Service
}

func setupSessionModule(t *testing.T) *sessionModule {
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

	client := setupDynamo(t, sessionTable, identityTable)

	authenticationIdentityStore := &dynamodbadapters.AuthNIdentityRepository{
		Client:    client,
		Table:     aws.String("test_identity_table"),
		NameIndex: aws.String("name-index"),
	}

	sessionStore := &dynamodbadapters.SessionRepository{
		Client:    client,
		Table:     aws.String("test_session_table"),
		NameIndex: aws.String("secondary-lookup-index"),
	}

	sessionService := authentication.SessionService{
		SessionStore:  sessionStore,
		IdentityStore: authenticationIdentityStore,
	}

	identityService := identity.Service{
		Repo: &dynamodbadapters.IdentityRepository{
			Client: client,
			Table:  aws.String("test_identity_table"),
		},
	}

	return &sessionModule{
		sessionService:              sessionService,
		identityService:             identityService,
		sessionStore:                sessionStore,
		authenticationIdentityStore: authenticationIdentityStore,
	}
}

func TestLoginUnknownUser(t *testing.T) {
	module := setupSessionModule(t)

	_, err := module.sessionService.Login(context.Background(), "unknown name", TEST_SECRET)
	assert.ErrorContains(t, err, "identity not found")
}

func TestLoginInvalidPassword(t *testing.T) {
	module := setupSessionModule(t)

	_, err := module.identityService.Register(context.Background(), TEST_NAME, TEST_SECRET)
	assert.NoError(t, err)

	_, err = module.sessionService.Login(context.Background(), TEST_NAME, "invalid secret")
	assert.ErrorContains(t, err, "invalid secret")
}

func TestLoginHappyPath(t *testing.T) {
	module := setupSessionModule(t)

	ident, err := module.identityService.Register(context.Background(), TEST_NAME, TEST_SECRET)
	assert.NoError(t, err)

	session, err := module.sessionService.Login(context.Background(), TEST_NAME, TEST_SECRET)
	assert.NoError(t, err)

	assert.NotEmpty(t, session.Id)
	assert.Equal(t, string(ident.Id), string(session.Owner))
}

func TestAuthenticateBySessionInvalidSessionId(t *testing.T) {
	module := setupSessionModule(t)

	_, err := module.sessionService.AuthenticateBySession(context.Background(), "")
	assert.ErrorContains(t, err, "session id cannot be null or empty")

	_, err = module.sessionService.AuthenticateBySession(context.Background(), "unknown session")
	assert.ErrorContains(t, err, "session expired")
}

func TestAuthenticateBySessionHappyPath(t *testing.T) {
	module := setupSessionModule(t)

	ident, err := module.identityService.Register(context.Background(), TEST_NAME, TEST_SECRET)
	assert.NoError(t, err)

	existingSession, err := module.sessionService.Login(context.Background(), TEST_NAME, TEST_SECRET)
	assert.NoError(t, err)

	session, err := module.sessionService.AuthenticateBySession(context.Background(), existingSession.Id)
	assert.NoError(t, err)

	assert.Equal(t, existingSession.Id, session.Id)
	assert.Equal(t, string(ident.Id), string(session.Owner))
}
