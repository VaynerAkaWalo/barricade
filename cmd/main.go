package main

import (
	handlers "barricade/internal/adapters/http"
	"barricade/internal/domain/healthcheck"
	"barricade/internal/infrastructure/server"
	"log"
)

func main() {
	healthHandler := handlers.HealthHttpHandler{
		Service: &healthcheck.Service{},
	}

	httpServer := &server.HttpServer{
		Addr:     ":8000",
		Handlers: []server.HttpRouteHandler{&healthHandler},
	}

	log.Fatal(httpServer.ListenAndServe())
}
