package dynamodbadapters

import (
	"barricade/internal/domain/authentication"
	"barricade/internal/infrastructure/htp"
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"log/slog"
	"net/http"
	"os"
)

type dbSessionAdapter struct {
	Id           string `dynamodbav:"id"`
	Owner        string `dynamodbav:"secondary-lookup"`
	ResourceType string `dynamodbav:"resource-type"`
	CreatedAt    int64  `dynamodbav:"createdAt"`
	ExpireAt     int64  `dynamodbav:"expireAt"`
}

type SessionRepository struct {
	Client    *dynamodb.Client
	Table     *string
	NameIndex *string
}

func NewSessionRepository(cfg aws.Config) *SessionRepository {
	return &SessionRepository{
		Client:    dynamodb.NewFromConfig(cfg),
		Table:     aws.String(os.Getenv("SESSION_TABLE")),
		NameIndex: aws.String(os.Getenv("SESSION_TABLE_NAME_INDEX")),
	}
}

func (r *SessionRepository) Save(ctx context.Context, session *authentication.Session) error {
	dbSession := &dbSessionAdapter{
		Id:           string(session.Id),
		Owner:        string(session.Owner),
		ResourceType: "session-v1",
		CreatedAt:    session.CreatedAt,
		ExpireAt:     session.ExpireAt,
	}

	item, err := attributevalue.MarshalMap(dbSession)
	if err != nil {
		return err
	}

	_, err = r.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: r.Table,
		Item:      item,
	})

	return err
}

func (r *SessionRepository) FindById(ctx context.Context, id authentication.SessionId) (*authentication.Session, error) {
	key, _ := attributevalue.Marshal(id)

	output, err := r.Client.GetItem(ctx, &dynamodb.GetItemInput{
		Key:            map[string]types.AttributeValue{"id": key},
		TableName:      r.Table,
		ConsistentRead: aws.Bool(false),
	})

	if err != nil {
		return nil, err
	}

	if len(output.Item) == 0 {
		return nil, htp.NewError("session not found", http.StatusNotFound)
	}

	var item dbSessionAdapter
	err = attributevalue.UnmarshalMap(output.Item, &item)
	if err != nil {
		return nil, err
	}

	return &authentication.Session{
		Id:        id,
		Owner:     authentication.IdentityId(item.Owner),
		CreatedAt: item.CreatedAt,
		ExpireAt:  item.ExpireAt,
	}, nil
}

func (r *SessionRepository) FindByIdentity(ctx context.Context, ownerId authentication.IdentityId) (*authentication.Session, error) {
	keyEx := expression.Key("secondary-lookup").Equal(expression.Value(ownerId))
	expr, err := expression.NewBuilder().WithKeyCondition(keyEx).Build()
	if err != nil {
		return nil, htp.NewError("cannot construct key", http.StatusInternalServerError)
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
		return nil, htp.NewError("error while looking for session", http.StatusInternalServerError)
	}

	if len(output.Items) == 0 {
		return nil, htp.NewError("session not found", http.StatusNotFound)
	}

	if len(output.Items) > 1 {
		slog.ErrorContext(ctx, fmt.Sprintf("found %d open sessions for identity %s", len(output.Items), ownerId))
	}

	var session dbSessionAdapter
	err = attributevalue.UnmarshalMap(output.Items[0], &session)
	if err != nil {
		return nil, htp.NewError("db result does not satisfy required schema", http.StatusInternalServerError)
	}

	return &authentication.Session{
		Id:        authentication.SessionId(session.Id),
		Owner:     authentication.IdentityId(session.Owner),
		CreatedAt: session.CreatedAt,
		ExpireAt:  session.ExpireAt,
	}, nil
}
