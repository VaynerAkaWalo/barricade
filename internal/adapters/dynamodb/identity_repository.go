package dynamodbadapters

import (
	"barricade/internal/domain/identity"
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"os"
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
	Client *dynamodb.Client
	Table  *string
}

func NewIdentityRepository(cfg aws.Config) *IdentityRepository {
	return &IdentityRepository{
		Client: dynamodb.NewFromConfig(cfg),
		Table:  aws.String(os.Getenv("IDENTITY_TABLE_NAME")),
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
