package itest

import (
	"context"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/stretchr/testify/assert"
	tcdynamodb "github.com/testcontainers/testcontainers-go/modules/dynamodb"
)

var (
	sharedHost string
	startOnce  sync.Once
)

func SetupDynamo(t *testing.T, tables ...dynamodb.CreateTableInput) *dynamodb.Client {
	startOnce.Do(func() {
		ctx := context.Background()

		container, err := tcdynamodb.Run(ctx, "amazon/dynamodb-local:latest")
		if err != nil {
			log.Fatalf("failed to start DynamoDB container: %v", err)
		}

		sharedHost, err = container.ConnectionString(ctx)
		if err != nil {
			log.Fatalf("failed to get DynamoDB connection string: %v", err)
		}
	})

	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
		Value: aws.Credentials{
			AccessKeyID:     "dummy",
			SecretAccessKey: "dummysecret",
		},
	}))
	assert.NoError(t, err)

	client := dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		o.BaseEndpoint = aws.String("http://" + sharedHost)
	})

	assert.EventuallyWithT(t, func(collect *assert.CollectT) {
		for _, table := range tables {
			_, err = client.CreateTable(context.Background(), &table)
			assert.NoError(collect, err, "waiting for table creation")
		}
	}, 30*time.Second, 1*time.Second, "Failed to create table after retries")

	t.Cleanup(func() {
		for _, table := range tables {
			_, err := client.DeleteTable(context.Background(), &dynamodb.DeleteTableInput{
				TableName: table.TableName,
			})
			if err != nil {
				log.Printf("failed to delete table %s: %v", *table.TableName, err)
			}
		}
	})

	return client
}
