package identity

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	tcdynamodb "github.com/testcontainers/testcontainers-go/modules/dynamodb"
	"golang.org/x/crypto/bcrypt"
)

const (
	TEST_NAME   = "first name"
	TEST_SECRET = "changeIt"
)

type identityModule struct {
	service    Service
	repository Repository
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

	assert.EventuallyWithT(t, func(c *assert.CollectT) {
		_, err = client.CreateTable(ctx, &dynamodb.CreateTableInput{
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
				{
					AttributeName: aws.String("sharded-type"),
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
		})
		assert.NoError(c, err, "waiting for table creation")
	}, 30*time.Second, 1*time.Second, "Failed to create table after retries")

	ddbRepository := DynamoDBIdentityRepository{
		Client:               client,
		Table:                aws.String("test_identity_table"),
		SecondaryLookupIndex: aws.String("secondary-lookup-index"),
		ShardedTypeIndex:     aws.String("sharded-type-index"),
	}

	return &identityModule{
		service:    Service{Repo: &ddbRepository},
		repository: &ddbRepository,
	}
}

func TestRegisterInputValidation(t *testing.T) {
	module := setupModule(t)

	_, err := module.service.Register(context.Background(), "", TEST_SECRET)
	assert.ErrorIs(t, err, ErrEmptyName)

	_, err = module.service.Register(context.Background(), TEST_NAME, "")
	assert.ErrorIs(t, err, ErrEmptySecret)
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

func TestIdentityFindAllHappyPath(t *testing.T) {
	module := setupModule(t)

	_, err := module.service.Register(context.Background(), "user-1", TEST_SECRET)
	assert.NoError(t, err)

	_, err = module.service.Register(context.Background(), "user-2", TEST_SECRET)
	assert.NoError(t, err)

	_, err = module.service.Register(context.Background(), "user-3", TEST_SECRET)
	assert.NoError(t, err)

	identities, err := module.service.FindAll(context.Background())
	assert.NoError(t, err)
	assert.Len(t, identities, 3)
}
