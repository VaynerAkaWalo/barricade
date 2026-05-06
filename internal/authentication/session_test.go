package authentication

import (
	"context"
	"testing"

	"barricade/internal/identity"
	"barricade/internal/itest"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
)

const (
	TEST_NAME   = "first name"
	TEST_SECRET = "changeIt"
)

type sessionModule struct {
	sessionService        SessionService
	sessionStore          SessionRepository
	authenticationService Service
	authenticationStore   IdentityRepository
	identityService       identity.Service
}

func setupSessionModule(t *testing.T) *sessionModule {
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

	client := itest.SetupDynamo(t, sessionTable, identityTable)

	identityStore := &identity.DynamoDBIdentityRepository{
		Client:               client,
		Table:                aws.String("test_identity_table"),
		SecondaryLookupIndex: aws.String("secondary-lookup-index"),
	}

	sessionStore := &DynamoDBSessionRepository{
		Client:               client,
		Table:                aws.String("test_session_table"),
		SecondaryLookupIndex: aws.String("secondary-lookup-index"),
	}

	sessionService := SessionService{
		SessionStore:  sessionStore,
		IdentityStore: identityStore,
	}

	identityService := identity.Service{
		Repo: &identity.DynamoDBIdentityRepository{
			Client: client,
			Table:  aws.String("test_identity_table"),
		},
	}

	authenticationService := Service{
		IdentityStore: identityStore,
		SessionStore:  sessionStore,
	}

	return &sessionModule{
		sessionService:        sessionService,
		identityService:       identityService,
		sessionStore:          sessionStore,
		authenticationStore:   identityStore,
		authenticationService: authenticationService,
	}
}

func TestLoginUnknownUser(t *testing.T) {
	module := setupSessionModule(t)

	_, err := module.sessionService.CreateOrGetSessionForCredentials(context.Background(), "unknown name", TEST_SECRET)
	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestLoginInvalidPassword(t *testing.T) {
	module := setupSessionModule(t)

	_, err := module.identityService.Register(context.Background(), TEST_NAME, TEST_SECRET)
	assert.NoError(t, err)

	_, err = module.sessionService.CreateOrGetSessionForCredentials(context.Background(), TEST_NAME, "invalid secret")
	assert.ErrorIs(t, err, ErrInvalidCredentials)
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
	assert.ErrorIs(t, err, ErrEmptySessionId)

	_, err = module.authenticationService.AuthenticateBySession(context.Background(), "unknown session")
	assert.ErrorIs(t, err, ErrSessionNotFound)
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
