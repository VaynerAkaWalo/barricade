package main

import (
	"barricade/internal/authentication"
	"barricade/internal/identity"
	"barricade/internal/infrastructure"
	"barricade/internal/infrastructure/ihttp"
	"context"
	"log"
	"log/slog"

	"github.com/VaynerAkaWalo/go-toolkit/xhttp"
	"github.com/VaynerAkaWalo/go-toolkit/xlog"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/caarlos0/env/v11"
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

	identityRepository := identity.NewIdentityRepository(awsCfg)

	identityHandler := identity.HttpHandler{
		Service: identity.Service{
			Repo: identityRepository,
		},
	}

	sessionService := authentication.SessionService{
		SessionStore:  authentication.NewSessionRepository(awsCfg),
		IdentityStore: identityRepository,
	}

	authNHandler := authentication.HttpHandler{
		Service:     sessionService,
		Domain:      cfg.Domain,
		SessionTime: cfg.SessionTime,
	}

	authNService := authentication.Service{
		IdentityStore: identityRepository,
		SessionStore:  sessionService.SessionStore,
	}

	authenticator := xhttp.NewAuthenticator(
		ihttp.BarricadeAuthenticationProvider{
			AuthenticationService: authNService,
		},
		[]string{"GET /health", "POST /v1/login", "POST /v1/register"}...)

	httpServer := xhttp.Server{
		Addr:     ":8080",
		Handlers: []xhttp.RouteHandler{&identityHandler, &authNHandler, &infrastructure.HealthHttpHandler{}},
		AuthN:    authenticator,
	}

	slog.Error(httpServer.ListenAndServe().Error())
}
