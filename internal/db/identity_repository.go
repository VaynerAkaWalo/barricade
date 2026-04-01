package db

import (
	"barricade/internal/identity"
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type identityDDB struct {
	Id         string `dynamodbav:"id"`
	Name       string `dynamodbav:"name"`
	SecretHash []byte `dynamodbav:"secret"`
	CreatedAt  int64  `dynamodbav:"createdAt"`
	UpdatedAt  int64  `dynamodbav:"updatedAt"`
}

func convertToDB(identity *identity.Identity) *identityDDB {
	return &identityDDB{
		Id:         string(identity.Id),
		Name:       identity.Name,
		SecretHash: identity.SecretHash,
		CreatedAt:  identity.CreatedAt,
		UpdatedAt:  identity.UpdatedAt,
	}
}

func key(id identity.Id) map[string]types.AttributeValue {
	dbId, _ := attributevalue.Marshal(id)
	return map[string]types.AttributeValue{"id": dbId}
}

type IdentityRepository struct {
	Client    *dynamodb.Client
	Table     *string
	NameIndex *string
}

func NewIdentityRepository(cfg aws.Config) *IdentityRepository {
	return &IdentityRepository{
		Client:    dynamodb.NewFromConfig(cfg),
		Table:     aws.String(os.Getenv("IDENTITY_TABLE_NAME")),
		NameIndex: aws.String("name-index"),
	}
}

func (r *IdentityRepository) Save(ctx context.Context, identity *identity.Identity) error {
	item, err := attributevalue.MarshalMap(convertToDB(identity))
	if err != nil {
		return err
	}

	_, err = r.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: r.Table,
		Item:      item,
	})

	if err != nil {
		return err
	}

	return nil
}

func (r *IdentityRepository) FindById(ctx context.Context, id identity.Id) (*identity.Identity, error) {
	output, err := r.Client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName:      r.Table,
		ConsistentRead: aws.Bool(false),
		Key:            key(id),
	})

	if err != nil {
		return nil, err
	}

	if len(output.Item) == 0 {
		return nil, identity.ErrIdentityNotFound
	}

	var dbEntity identityDDB

	err = attributevalue.UnmarshalMap(output.Item, &dbEntity)
	if err != nil {
		return nil, err
	}

	entity := &identity.Identity{
		Id:         id,
		Name:       dbEntity.Name,
		SecretHash: dbEntity.SecretHash,
		CreatedAt:  dbEntity.CreatedAt,
		UpdatedAt:  dbEntity.UpdatedAt,
	}

	return entity, nil
}

func (r *IdentityRepository) FindByName(ctx context.Context, name string) (*identity.Identity, error) {
	keyEx := expression.Key("name").Equal(expression.Value(name))
	expr, err := expression.NewBuilder().WithKeyCondition(keyEx).Build()
	if err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("failed to build expression: %v", err))
		return nil, err
	}

	output, err := r.Client.Query(ctx, &dynamodb.QueryInput{
		TableName:                 r.Table,
		IndexName:                 r.NameIndex,
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	})
	if err != nil {
		slog.ErrorContext(ctx, err.Error())
		return nil, err
	}

	if len(output.Items) == 0 {
		return nil, identity.ErrIdentityNotFound
	}

	if len(output.Items) > 1 {
		slog.ErrorContext(ctx, fmt.Sprintf("found %d identities with name %s", len(output.Items), name))
		return nil, identity.ErrDuplicateIdentityName
	}

	var entity identityDDB
	err = attributevalue.UnmarshalMap(output.Items[0], &entity)
	if err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("failed to unmarshal identity: %v", err))
		return nil, err
	}

	return &identity.Identity{
		Id:         identity.Id(entity.Id),
		Name:       entity.Name,
		SecretHash: entity.SecretHash,
	}, nil
}
