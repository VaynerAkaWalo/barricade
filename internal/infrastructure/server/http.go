package server

import (
	"log"
	"net/http"
)

type HttpServer struct {
	Addr     string
	Handlers []HttpRouteHandler
}

type HttpRouteHandler interface {
	RegisterRoutes(router *http.ServeMux)
}

func (server *HttpServer) ListenAndServe() error {
	router := http.NewServeMux()

	httpServer := &http.Server{
		Addr:    server.Addr,
		Handler: router,
	}

	for _, handler := range server.Handlers {
		handler.RegisterRoutes(router)
	}

	log.Println("Starting http server at port " + server.Addr)
	return httpServer.ListenAndServe()
}
