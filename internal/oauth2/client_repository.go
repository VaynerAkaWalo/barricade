package oauth2

import (
	"context"
	"hash/crc32"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const shardCount = 4

type clientDDB struct {
	Id                string `dynamodbav:"id"`
	Type              string `dynamodbav:"type"`
	ShardedType       string `dynamodbav:"sharded-type"`
	SecondaryLookup   string `dynamodbav:"secondary-lookup"`
	SecondaryLookupSk string `dynamodbav:"secondary-lookup-sk"`
	Name              string `dynamodbav:"name"`
	Domain            string `dynamodbav:"domain"`
	SecretHash        []byte `dynamodbav:"secret"`
	RedirectURI       string `dynamodbav:"redirectURI"`
	CreatedAt         int64  `dynamodbav:"createdAt"`
	UpdatedAt         int64  `dynamodbav:"updatedAt"`
}

func computeShard(id ClientId) string {
	sum := crc32.ChecksumIEEE([]byte(id))
	return "oauth-client#" + strconv.Itoa(int(sum)%shardCount)
}

func convertClientToDB(c *Client) *clientDDB {
	return &clientDDB{
		Id:                string(c.Id),
		Type:              "oauth-client",
		ShardedType:       computeShard(c.Id),
		SecondaryLookup:   c.OwnerId,
		SecondaryLookupSk: "oauth-client",
		Name:              c.Name,
		Domain:            c.Domain,
		SecretHash:        c.SecretHash,
		RedirectURI:       c.RedirectURI,
		CreatedAt:         c.CreatedAt,
		UpdatedAt:         c.UpdatedAt,
	}
}

func convertClientFromDB(db *clientDDB) *Client {
	return &Client{
		Id:          ClientId(db.Id),
		OwnerId:     db.SecondaryLookup,
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
	Client           *dynamodb.Client
	Table            *string
	ShardedTypeIndex *string
}

func NewClientRepository(cfg aws.Config) *DynamoDBClientRepository {
	return &DynamoDBClientRepository{
		Client:           dynamodb.NewFromConfig(cfg),
		Table:            aws.String(os.Getenv("CLIENT_TABLE_NAME")),
		ShardedTypeIndex: aws.String("sharded-type-index"),
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

func (r *DynamoDBClientRepository) FindAll(ctx context.Context) ([]*Client, error) {
	var result []*Client

	for i := 0; i < shardCount; i++ {
		shard := "oauth-client#" + strconv.Itoa(i)
		keyEx := expression.Key("sharded-type").Equal(expression.Value(shard))
		expr, err := expression.NewBuilder().WithKeyCondition(keyEx).Build()
		if err != nil {
			return nil, err
		}

		output, err := r.Client.Query(ctx, &dynamodb.QueryInput{
			TableName:                 r.Table,
			IndexName:                 r.ShardedTypeIndex,
			KeyConditionExpression:    expr.KeyCondition(),
			ExpressionAttributeNames:  expr.Names(),
			ExpressionAttributeValues: expr.Values(),
		})
		if err != nil {
			return nil, err
		}

		for _, item := range output.Items {
			var dbEntity clientDDB
			err = attributevalue.UnmarshalMap(item, &dbEntity)
			if err != nil {
				return nil, err
			}
			result = append(result, convertClientFromDB(&dbEntity))
		}
	}

	return result, nil
}
