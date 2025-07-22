package main

import (
	dynamodbadapters "barricade/internal/adapters/dynamodb"
	handlers "barricade/internal/adapters/http"
	"barricade/internal/domain/authentication"
	"barricade/internal/domain/identity"
	"barricade/internal/infrastructure/htp"
	"barricade/internal/infrastructure/logging"
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"log"
	"log/slog"
	"os"
)

func main() {
	handler := logging.ContextHandler{Handler: slog.NewJSONHandler(os.Stdout, nil)}
	slog.SetDefault(slog.New(handler))

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

	authNHandler := handlers.AuthenticationHttpHandler{
		Service: authentication.SessionService{
			SessionStore:  dynamodbadapters.NewSessionRepository(awsCfg),
			IdentityStore: dynamodbadapters.NewAuthNIdentityRepository(awsCfg),
		},
	}

	httpServer := &htp.Server{
		Addr:     ":8000",
		Handlers: []htp.RouteHandler{&handlers.HealthHttpHandler{}, &identityHandler, &authNHandler},
	}

	log.Fatal(httpServer.ListenAndServe())
}
