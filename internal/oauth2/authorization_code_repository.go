package oauth2

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type authCodeDDB struct {
	Id                  string `dynamodbav:"id"`
	Type                string `dynamodbav:"type"`
	SecondaryLookup     string `dynamodbav:"secondary-lookup"`
	SecondaryLookupSk   string `dynamodbav:"secondary-lookup-sk"`
	ClientId            string `dynamodbav:"clientId"`
	RedirectURI         string `dynamodbav:"redirectURI"`
	Scope               string `dynamodbav:"scope"`
	Nonce               string `dynamodbav:"nonce,omitempty"`
	State               string `dynamodbav:"state,omitempty"`
	CodeChallenge       string `dynamodbav:"codeChallenge,omitempty"`
	CodeChallengeMethod string `dynamodbav:"codeChallengeMethod,omitempty"`
	AuthTime            int64  `dynamodbav:"authTime"`
	CreatedAt           int64  `dynamodbav:"createdAt"`
	ExpireAt            int64  `dynamodbav:"expireAt"`
}

type AuthorizationCodeRepository interface {
	Save(ctx context.Context, code *AuthorizationCode) error
	FindByCode(ctx context.Context, code string) (*AuthorizationCode, error)
	Delete(ctx context.Context, code string) error
}

type DynamoDBAuthorizationCodeRepository struct {
	Client               *dynamodb.Client
	Table                *string
	SecondaryLookupIndex *string
}

func NewAuthorizationCodeRepository(cfg aws.Config) *DynamoDBAuthorizationCodeRepository {
	return &DynamoDBAuthorizationCodeRepository{
		Client:               dynamodb.NewFromConfig(cfg),
		Table:                aws.String(os.Getenv("OPERATIONAL_TABLE")),
		SecondaryLookupIndex: aws.String("secondary-lookup-index"),
	}
}

func (r *DynamoDBAuthorizationCodeRepository) Save(ctx context.Context, code *AuthorizationCode) error {
	dbCode := &authCodeDDB{
		Id:                  code.Code,
		Type:                "authorization-code",
		SecondaryLookup:     code.IdentityId,
		SecondaryLookupSk:   "authorization-code",
		ClientId:            code.ClientId,
		RedirectURI:         code.RedirectURI,
		Scope:               code.Scope,
		Nonce:               code.Nonce,
		State:               code.State,
		CodeChallenge:       code.CodeChallenge,
		CodeChallengeMethod: code.CodeChallengeMethod,
		AuthTime:            code.AuthTime,
		CreatedAt:           code.CreatedAt,
		ExpireAt:            code.ExpireAt,
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
	dbCodeId, _ := attributevalue.Marshal(code)
	dbType, _ := attributevalue.Marshal("authorization-code")

	output, err := r.Client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName:      r.Table,
		Key:            map[string]types.AttributeValue{"id": dbCodeId, "type": dbType},
		ConsistentRead: aws.Bool(false),
	})
	if err != nil {
		return nil, err
	}

	if len(output.Item) == 0 {
		return nil, ErrInvalidCode
	}

	var dbCode authCodeDDB
	err = attributevalue.UnmarshalMap(output.Item, &dbCode)
	if err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("failed to unmarshal authorization code: %v", err))
		return nil, err
	}

	return &AuthorizationCode{
		Code:                dbCode.Id,
		ClientId:            dbCode.ClientId,
		IdentityId:          dbCode.SecondaryLookup,
		RedirectURI:         dbCode.RedirectURI,
		Scope:               dbCode.Scope,
		Nonce:               dbCode.Nonce,
		State:               dbCode.State,
		CodeChallenge:       dbCode.CodeChallenge,
		CodeChallengeMethod: dbCode.CodeChallengeMethod,
		AuthTime:            dbCode.AuthTime,
		CreatedAt:           dbCode.CreatedAt,
		ExpireAt:            dbCode.ExpireAt,
	}, nil
}

func (r *DynamoDBAuthorizationCodeRepository) Delete(ctx context.Context, code string) error {
	dbCodeId, _ := attributevalue.Marshal(code)
	dbType, _ := attributevalue.Marshal("authorization-code")

	_, err := r.Client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: r.Table,
		Key: map[string]types.AttributeValue{
			"id":   dbCodeId,
			"type": dbType,
		},
	})

	return err
}
