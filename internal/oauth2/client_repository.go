package oauth2

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type clientDDB struct {
	Id          string `dynamodbav:"id"`
	Type        string `dynamodbav:"type"`
	Name        string `dynamodbav:"name"`
	Domain      string `dynamodbav:"domain"`
	SecretHash  []byte `dynamodbav:"secret"`
	RedirectURI string `dynamodbav:"redirectURI"`
	CreatedAt   int64  `dynamodbav:"createdAt"`
	UpdatedAt   int64  `dynamodbav:"updatedAt"`
}

func convertClientToDB(c *Client) *clientDDB {
	return &clientDDB{
		Id:          string(c.Id),
		Type:        "oauth-client",
		Name:        c.Name,
		Domain:      c.Domain,
		SecretHash:  c.SecretHash,
		RedirectURI: c.RedirectURI,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}

func convertClientFromDB(db *clientDDB) *Client {
	return &Client{
		Id:          ClientId(db.Id),
		Name:        db.Name,
		Domain:      db.Domain,
		SecretHash:  db.SecretHash,
		RedirectURI: db.RedirectURI,
		CreatedAt:   db.CreatedAt,
		UpdatedAt:   db.UpdatedAt,
	}
}

func clientKey(id ClientId) map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"id":   &types.AttributeValueMemberS{Value: string(id)},
		"type": &types.AttributeValueMemberS{Value: "oauth-client"},
	}
}

type DynamoDBClientRepository struct {
	Client *dynamodb.Client
	Table  *string
}

func NewClientRepository(cfg aws.Config) *DynamoDBClientRepository {
	return &DynamoDBClientRepository{
		Client: dynamodb.NewFromConfig(cfg),
		Table:  aws.String(os.Getenv("CLIENT_TABLE_NAME")),
	}
}

func (r *DynamoDBClientRepository) Save(ctx context.Context, c *Client) error {
	item, err := attributevalue.MarshalMap(convertClientToDB(c))
	if err != nil {
		return err
	}

	_, err = r.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: r.Table,
		Item:      item,
	})

	return err
}

func (r *DynamoDBClientRepository) FindById(ctx context.Context, id ClientId) (*Client, error) {
	output, err := r.Client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName:      r.Table,
		ConsistentRead: aws.Bool(false),
		Key:            clientKey(id),
	})
	if err != nil {
		return nil, err
	}

	if len(output.Item) == 0 {
		return nil, ErrClientNotFound
	}

	var dbEntity clientDDB
	err = attributevalue.UnmarshalMap(output.Item, &dbEntity)
	if err != nil {
		return nil, err
	}

	return convertClientFromDB(&dbEntity), nil
}
