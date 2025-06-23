package main

import (
	dynamodbadapters "barricade/internal/adapters/dynamodb"
	handlers "barricade/internal/adapters/http"
	"barricade/internal/domain/healthcheck"
	"barricade/internal/domain/identity"
	"barricade/internal/infrastructure/httpserver"
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"log"
	"os"
)

func main() {
	healthHandler := handlers.HealthHttpHandler{
		Service: &healthcheck.Service{},
	}

	cp := credentials.NewStaticCredentialsProvider(os.Getenv("DDB_ACCESS_KEY"), os.Getenv("DDB_ACCESS_SECRET_KEY"), "")

	awsCfg, err := config.LoadDefaultConfig(context.TODO(), config.WithCredentialsProvider(cp), config.WithRegion("eu-north-1"))
	if err != nil {
		log.Fatal(err)
	}

	identityHandler := handlers.IdentityHttpHandler{
		Service: identity.Service{
			Repo: dynamodbadapters.NewIdentityRepository(awsCfg),
		},
	}

	httpServer := &httpserver.Server{
		Addr:     ":8000",
		Handlers: []httpserver.RouteHandler{&healthHandler, &identityHandler},
	}

	log.Fatal(httpServer.ListenAndServe())
}
