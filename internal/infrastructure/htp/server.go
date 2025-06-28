package htp

import (
	"log"
	"net/http"
)

type Server struct {
	Addr     string
	Handlers []RouteHandler
}

type RouteHandler interface {
	RegisterRoutes(router *http.ServeMux)
}

func (server *Server) ListenAndServe() error {
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
