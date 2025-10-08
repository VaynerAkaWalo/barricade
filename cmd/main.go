package main

import (
	dynamodbadapters "barricade/internal/adapters/dynamodb"
	handlers "barricade/internal/adapters/http"
	"barricade/internal/domain/authentication"
	"barricade/internal/domain/identity"
	"context"
	"github.com/VaynerAkaWalo/go-toolkit/xhttp"
	"github.com/VaynerAkaWalo/go-toolkit/xlog"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"log"
	"log/slog"
	"os"
)

const (
	DomainEnv = "DOMAIN"
)

func main() {
	slog.SetDefault(slog.New(xlog.NewPreConfiguredHandler()))

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
		Domain: os.Getenv(DomainEnv),
	}

	httpServer := xhttp.Server{
		Addr:     ":8080",
		Handlers: []xhttp.RouteHandler{&identityHandler, &authNHandler, &handlers.HealthHttpHandler{}},
	}

	slog.Error(httpServer.ListenAndServe().Error())
}
