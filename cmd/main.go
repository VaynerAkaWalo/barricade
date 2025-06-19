package main

import (
	handlers "barricade/internal/adapters/http"
	"barricade/internal/domain/healthcheck"
	"barricade/internal/infrastructure/httpserver"
	"log"
)

func main() {
	healthHandler := handlers.HealthHttpHandler{
		Service: &healthcheck.Service{},
	}

	httpServer := &httpserver.Server{
		Addr:     ":8000",
		Handlers: []httpserver.RouteHandler{&healthHandler},
	}

	log.Fatal(httpServer.ListenAndServe())
}
