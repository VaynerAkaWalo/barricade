package main

import (
	"barricade/internal/authentication"
	"barricade/internal/identity"
	"barricade/internal/infrastructure"
	"barricade/internal/infrastructure/ihttp"
	"barricade/internal/keys"
	"barricade/internal/oauth2"
	"barricade/internal/oidc"
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
	Domain             string `env:"DOMAIN"`
	SessionTime        int    `env:"SESSION_TIME" envDefault:"7200"`
	AwsAccessKey       string `env:"DDB_ACCESS_KEY"`
	AwsSecretKey       string `env:"DDB_ACCESS_SECRET_KEY"`
	IssuerURL          string `env:"ISSUER_URL" envDefault:"https://auth.blamedevs.com"`
	LoginURL           string `env:"LOGIN_URL" envDefault:"https://auth.blamedevs.com/login"`
	TokenExpiryMinutes int    `env:"TOKEN_EXPIRY_MINUTES" envDefault:"5"`
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

	keyRepo := keys.NewInMemoryRepository()
	keyService := keys.NewService(keyRepo)

	_, err = keyService.CreateKey(context.TODO(), keys.RS256)
	if err != nil {
		log.Fatal("failed to create initial signing key: ", err)
	}

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

	jwksHandler := oidc.JWKSHandler{
		KeyService: keyService,
	}

	authorizeService := oauth2.AuthorizeService{
		IdentityStore: identityRepository,
		KeyService:    keyService,
		Issuer:        cfg.IssuerURL,
		TokenExpiry:   cfg.TokenExpiryMinutes,
	}

	authorizeHandler := oauth2.HttpHandler{
		Service:            &authorizeService,
		LoginURL:           cfg.LoginURL,
		DefaultRedirectURI: cfg.IssuerURL,
	}

	authenticator := xhttp.NewAuthenticator(
		ihttp.BarricadeAuthenticationProvider{
			AuthenticationService: authNService,
		},
		[]string{"GET /health", "POST /v1/login", "POST /v1/register", "GET /.well-known/jwks.json", "GET /v1/oauth2/authorize"}...)

	httpServer := xhttp.Server{
		Addr:     ":8080",
		Handlers: []xhttp.RouteHandler{&identityHandler, &authNHandler, &infrastructure.HealthHttpHandler{}, &jwksHandler, &authorizeHandler},
		AuthN:    authenticator,
	}

	slog.Error(httpServer.ListenAndServe().Error())
}
