package dynamodbadapters

import (
	"barricade/internal/domain/authentication"
	"context"
	"fmt"
	"github.com/VaynerAkaWalo/go-toolkit/xhttp"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"log/slog"
	"net/http"
	"os"
)

type authNIdentityAdapter struct {
	Id         string `dynamodbav:"id"`
	Name       string `dynamodbav:"name"`
	SecretHash []byte `dynamodbav:"secret"`
}

type AuthNIdentityRepository struct {
	Client    *dynamodb.Client
	Table     *string
	NameIndex *string
}

func NewAuthNIdentityRepository(cfg aws.Config) *AuthNIdentityRepository {
	return &AuthNIdentityRepository{
		Client:    dynamodb.NewFromConfig(cfg),
		Table:     aws.String(os.Getenv("IDENTITY_TABLE_NAME")),
		NameIndex: aws.String("name-index"),
	}
}

func (r *AuthNIdentityRepository) FindByName(ctx context.Context, name string) (*authentication.Identity, error) {
	keyEx := expression.Key("name").Equal(expression.Value(name))
	expr, err := expression.NewBuilder().WithKeyCondition(keyEx).Build()
	if err != nil {
		return nil, xhttp.NewError("cannot construct key", http.StatusInternalServerError)
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
		return nil, xhttp.NewError("error occurred while looking for identity", http.StatusInternalServerError)
	}

	if len(output.Items) == 0 {
		return nil, xhttp.NewError("identity not found", http.StatusNotFound)
	}

	if len(output.Items) > 1 {
		slog.ErrorContext(ctx, fmt.Sprintf("found %d identities with name %s", len(output.Items), name))
		return nil, xhttp.NewError("duplicated name", http.StatusConflict)
	}

	var entity authNIdentityAdapter
	err = attributevalue.UnmarshalMap(output.Items[0], &entity)
	if err != nil {
		return nil, xhttp.NewError("error while serializing query result", http.StatusInternalServerError)
	}

	return &authentication.Identity{
		Id:         authentication.IdentityId(entity.Id),
		Name:       entity.Name,
		SecretHash: entity.SecretHash,
	}, nil
}
