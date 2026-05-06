package oauth2

import (
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

type authCodeDDB struct {
	Id              string `dynamodbav:"id"`
	Type            string `dynamodbav:"type"`
	SecondaryLookup string `dynamodbav:"secondary-lookup"`
	ResourceType    string `dynamodbav:"resource-type"`
	ClientId        string `dynamodbav:"clientId"`
	RedirectURI     string `dynamodbav:"redirectURI"`
	Scope           string `dynamodbav:"scope"`
	CreatedAt       int64  `dynamodbav:"createdAt"`
	ExpireAt        int64  `dynamodbav:"expireAt"`
}

type AuthorizationCodeRepository interface {
	Save(ctx context.Context, code *AuthorizationCode) error
	FindByCode(ctx context.Context, code string) (*AuthorizationCode, error)
	Delete(ctx context.Context, code string) error
}

type DynamoDBAuthorizationCodeRepository struct {
	Client    *dynamodb.Client
	Table     *string
	NameIndex *string
}

func NewAuthorizationCodeRepository(cfg aws.Config) *DynamoDBAuthorizationCodeRepository {
	return &DynamoDBAuthorizationCodeRepository{
		Client:    dynamodb.NewFromConfig(cfg),
		Table:     aws.String(os.Getenv("SESSION_TABLE")),
		NameIndex: aws.String(os.Getenv("SESSION_TABLE_NAME_INDEX")),
	}
}

func (r *DynamoDBAuthorizationCodeRepository) Save(ctx context.Context, code *AuthorizationCode) error {
	dbCode := &authCodeDDB{
		Id:              code.Code,
		Type:            "authorization-code",
		SecondaryLookup: code.IdentityId,
		ResourceType:    "authorization-code",
		ClientId:        code.ClientId,
		RedirectURI:     code.RedirectURI,
		Scope:           code.Scope,
		CreatedAt:       code.CreatedAt,
		ExpireAt:        code.ExpireAt,
	}

	item, err := attributevalue.MarshalMap(dbCode)
	if err != nil {
		return err
	}

	_, err = r.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: r.Table,
		Item:      item,
	})

	return err
}

func (r *DynamoDBAuthorizationCodeRepository) FindByCode(ctx context.Context, code string) (*AuthorizationCode, error) {
	keyEx := expression.Key("id").Equal(expression.Value(code))
	expr, err := expression.NewBuilder().WithKeyCondition(keyEx).Build()
	if err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("failed to build expression: %v", err))
		return nil, err
	}

	output, err := r.Client.Query(ctx, &dynamodb.QueryInput{
		TableName:                 r.Table,
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		Limit:                     aws.Int32(1),
	})
	if err != nil {
		return nil, err
	}

	if len(output.Items) == 0 {
		return nil, ErrInvalidCode
	}

	var dbCode authCodeDDB
	err = attributevalue.UnmarshalMap(output.Items[0], &dbCode)
	if err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("failed to unmarshal authorization code: %v", err))
		return nil, err
	}

	return &AuthorizationCode{
		Code:        dbCode.Id,
		ClientId:    dbCode.ClientId,
		IdentityId:  dbCode.SecondaryLookup,
		RedirectURI: dbCode.RedirectURI,
		Scope:       dbCode.Scope,
		CreatedAt:   dbCode.CreatedAt,
		ExpireAt:    dbCode.ExpireAt,
	}, nil
}

func (r *DynamoDBAuthorizationCodeRepository) Delete(ctx context.Context, code string) error {
	keyEx := expression.Key("id").Equal(expression.Value(code))
	expr, err := expression.NewBuilder().WithKeyCondition(keyEx).Build()
	if err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("failed to build expression: %v", err))
		return err
	}

	output, err := r.Client.Query(ctx, &dynamodb.QueryInput{
		TableName:                 r.Table,
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		Limit:                     aws.Int32(1),
	})
	if err != nil {
		return err
	}

	if len(output.Items) == 0 {
		return nil
	}

	var dbCode authCodeDDB
	err = attributevalue.UnmarshalMap(output.Items[0], &dbCode)
	if err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("failed to unmarshal authorization code: %v", err))
		return err
	}

	key := map[string]types.AttributeValue{
		"id": &types.AttributeValueMemberS{Value: dbCode.Id},
	}
	if dbCode.Type != "" {
		key["type"] = &types.AttributeValueMemberS{Value: dbCode.Type}
	}

	_, err = r.Client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: r.Table,
		Key:       key,
	})

	return err
}
