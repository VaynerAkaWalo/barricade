package test

import (
	"context"
	"testing"

	"barricade/internal/authentication"
	"barricade/internal/db"
	"barricade/internal/identity"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
)

type sessionModule struct {
	sessionService              authentication.SessionService
	sessionStore                authentication.SessionRepository
	authenticationService       authentication.Service
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

	identityStore := &db.IdentityRepository{
		Client:    client,
		Table:     aws.String("test_identity_table"),
		NameIndex: aws.String("name-index"),
	}

	sessionStore := &db.SessionRepository{
		Client:    client,
		Table:     aws.String("test_session_table"),
		NameIndex: aws.String("secondary-lookup-index"),
	}

	sessionService := authentication.SessionService{
		SessionStore:  sessionStore,
		IdentityStore: identityStore,
	}

	identityService := identity.Service{
		Repo: &db.IdentityRepository{
			Client: client,
			Table:  aws.String("test_identity_table"),
		},
	}

	authenticationService := authentication.Service{
		IdentityStore: identityStore,
		SessionStore:  sessionStore,
	}

	return &sessionModule{
		sessionService:              sessionService,
		identityService:             identityService,
		sessionStore:                sessionStore,
		authenticationIdentityStore: identityStore,
		authenticationService:       authenticationService,
	}
}

func TestLoginUnknownUser(t *testing.T) {
	module := setupSessionModule(t)

	_, err := module.sessionService.CreateOrGetSessionForCredentials(context.Background(), "unknown name", TEST_SECRET)
	assert.ErrorIs(t, err, authentication.ErrInvalidCredentials)
}

func TestLoginInvalidPassword(t *testing.T) {
	module := setupSessionModule(t)

	_, err := module.identityService.Register(context.Background(), TEST_NAME, TEST_SECRET)
	assert.NoError(t, err)

	_, err = module.sessionService.CreateOrGetSessionForCredentials(context.Background(), TEST_NAME, "invalid secret")
	assert.ErrorIs(t, err, authentication.ErrInvalidCredentials)
}

func TestLoginHappyPath(t *testing.T) {
	module := setupSessionModule(t)

	ident, err := module.identityService.Register(context.Background(), TEST_NAME, TEST_SECRET)
	assert.NoError(t, err)

	session, err := module.sessionService.CreateOrGetSessionForCredentials(context.Background(), TEST_NAME, TEST_SECRET)
	assert.NoError(t, err)

	assert.NotEmpty(t, session.Id)
	assert.Equal(t, string(ident.Id), string(session.Owner))
}

func TestAuthenticateBySessionInvalidSessionId(t *testing.T) {
	module := setupSessionModule(t)

	_, err := module.authenticationService.AuthenticateBySession(context.Background(), "")
	assert.ErrorIs(t, err, authentication.ErrEmptySessionId)

	_, err = module.authenticationService.AuthenticateBySession(context.Background(), "unknown session")
	assert.ErrorIs(t, err, authentication.ErrSessionNotFound)
}

func TestAuthenticateBySessionHappyPath(t *testing.T) {
	module := setupSessionModule(t)

	ident, err := module.identityService.Register(context.Background(), TEST_NAME, TEST_SECRET)
	assert.NoError(t, err)

	existingSession, err := module.sessionService.CreateOrGetSessionForCredentials(context.Background(), TEST_NAME, TEST_SECRET)
	assert.NoError(t, err)

	sessionOwner, err := module.authenticationService.AuthenticateBySession(context.Background(), existingSession.Id)
	assert.NoError(t, err)

	assert.Equal(t, string(ident.Id), string(sessionOwner.Id))
}
