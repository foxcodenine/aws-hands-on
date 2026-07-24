package httpserver

import (
	"03-dynamodb_crud_with_go/internal/api/routers"
	apptypes "03-dynamodb_crud_with_go/internal/app"
	"context"
	"fmt"
	"net/http"
)

type Server struct {
	HTTPServer *http.Server
	Port       string
}

func NewHttpServer(port string, app apptypes.App) *Server {

	mux := http.NewServeMux()
	mux.Handle("/", routers.Routes(app))

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: mux,
	}

	return &Server{
		HTTPServer: srv,
		Port:       port,
	}
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.HTTPServer.Shutdown(ctx)
}
