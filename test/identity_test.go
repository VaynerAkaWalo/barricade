package test

import (
	dynamodbadapters "barricade/internal/adapters/dynamodb"
	"barricade/internal/domain/identity"
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	tcdynamodb "github.com/testcontainers/testcontainers-go/modules/dynamodb"
	"golang.org/x/crypto/bcrypt"
	"log"
	"testing"
)

const (
	TEST_NAME   = "first name"
	TEST_SECRET = "changeIt"
)

type identityModule struct {
	service    identity.Service
	repository identity.Repository
}

func setupModule(t *testing.T) *identityModule {
	ctx := context.Background()

	ddb, err := tcdynamodb.Run(ctx, "amazon/dynamodb-local:latest")
	assert.NoError(t, err)

	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(ddb); err != nil {
			log.Println("failed to terminate container after tests")
		}
	})

	cfg, err := config.LoadDefaultConfig(ctx, config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
		Value: aws.Credentials{
			AccessKeyID:     "dummy",
			SecretAccessKey: "dummysecret",
		},
	}),
	)
	assert.NoError(t, err)

	host, err := ddb.ConnectionString(ctx)
	assert.NoError(t, err)

	fmt.Println(host)

	client := dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		o.BaseEndpoint = aws.String("http://" + host)
	})

	_, err = client.CreateTable(ctx, &dynamodb.CreateTableInput{
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
		},
		BillingMode: types.BillingModePayPerRequest,
	})
	assert.NoError(t, err)

	ddbRepository := dynamodbadapters.IdentityRepository{
		Client: client,
		Table:  aws.String("test_identity_table"),
	}

	return &identityModule{
		service:    identity.Service{Repo: &ddbRepository},
		repository: &ddbRepository,
	}
}

func TestRegisterInputValidation(t *testing.T) {
	module := setupModule(t)

	_, err := module.service.Register(context.Background(), "", TEST_SECRET)
	assert.ErrorContains(t, err, "name and secret cannot be null or empty")

	_, err = module.service.Register(context.Background(), TEST_NAME, "")
	assert.ErrorContains(t, err, "name and secret cannot be null or empty")
}

func TestRegisterHappyPath(t *testing.T) {
	module := setupModule(t)

	res, err := module.service.Register(context.Background(), TEST_NAME, TEST_SECRET)
	assert.NoError(t, err, "failed to create user")

	assert.Equal(t, TEST_NAME, res.Name)
	assert.Nil(t, bcrypt.CompareHashAndPassword(res.SecretHash, []byte(TEST_SECRET)))
	assert.NotEmpty(t, res.Id)
	assert.NotEmpty(t, res.UpdatedAt)
	assert.NotEmpty(t, res.CreatedAt)

	entity, err := module.repository.FindById(context.Background(), res.Id)
	assert.NoError(t, err)

	assert.Equal(t, res, entity)
}
