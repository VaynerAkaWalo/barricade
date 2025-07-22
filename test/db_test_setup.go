package test

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	tcdynamodb "github.com/testcontainers/testcontainers-go/modules/dynamodb"
	"log"
	"testing"
	"time"
)

func setupDynamo(t *testing.T, tables ...dynamodb.CreateTableInput) *dynamodb.Client {
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

	client := dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		o.BaseEndpoint = aws.String("http://" + host)
	})

	assert.EventuallyWithT(t, func(collect *assert.CollectT) {
		for _, table := range tables {
			_, err = client.CreateTable(ctx, &table)
			assert.NoError(collect, err, "waiting for table creation")
		}
	}, 30*time.Second, 1*time.Second, "Failed to create table after retries")

	return client
}
