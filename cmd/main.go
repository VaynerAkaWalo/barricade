package main

import (
	dynamodbadapters "barricade/internal/adapters/dynamodb"
	handlers "barricade/internal/adapters/http"
	"barricade/internal/domain/authentication"
	"barricade/internal/domain/identity"
	"barricade/internal/infrastructure/ihttp"
	"context"
	"github.com/VaynerAkaWalo/go-toolkit/xhttp"
	"github.com/VaynerAkaWalo/go-toolkit/xlog"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/caarlos0/env/v11"
	"log"
	"log/slog"
)

type appConfig struct {
	Domain       string `env:"DOMAIN"`
	SessionTime  int    `env:"SESSION_TIME" envDefault:"7200"`
	AwsAccessKey string `env:"DDB_ACCESS_KEY"`
	AwsSecretKey string `env:"DDB_ACCESS_SECRET_KEY"`
}

func main() {
	slog.SetDefault(slog.New(xlog.NewPreConfiguredHandler()))

	cfg, err := env.ParseAs[appConfig]()
	if err != nil {
		log.Fatal("unable to load env config")
	}

	cp := credentials.NewStaticCredentialsProvider(cfg.AwsAccessKey, cfg.AwsSecretKey, "")

	awsCfg, err := config.LoadDefaultConfig(context.TODO(), config.WithCredentialsProvider(cp), config.WithRegion("eu-north-1"))
	if err != nil {
		log.Fatal(err)
	}

	identityHandler := handlers.IdentityHttpHandler{
		Service: identity.Service{
			Repo: dynamodbadapters.NewIdentityRepository(awsCfg),
		},
	}

	sessionService := authentication.SessionService{
		SessionStore:  dynamodbadapters.NewSessionRepository(awsCfg),
		IdentityStore: dynamodbadapters.NewAuthNIdentityRepository(awsCfg),
	}

	authNHandler := handlers.AuthenticationHttpHandler{
		Service:     sessionService,
		Domain:      cfg.Domain,
		SessionTime: cfg.SessionTime,
	}

	authenticator := xhttp.NewAuthenticator(
		ihttp.BarricadeAuthenticationProvider{
			SessionService: sessionService,
		},
		[]string{"GET /health", "POST /v1/login", "POST /v1/register"}...)

	httpServer := xhttp.Server{
		Addr:     ":8080",
		Handlers: []xhttp.RouteHandler{&identityHandler, &authNHandler, &handlers.HealthHttpHandler{}},
		AuthN:    authenticator,
	}

	slog.Error(httpServer.ListenAndServe().Error())
}
