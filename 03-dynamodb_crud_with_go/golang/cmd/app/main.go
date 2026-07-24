package main

import (
	apptypes "03-dynamodb_crud_with_go/internal/app"
	"03-dynamodb_crud_with_go/internal/db"
	"03-dynamodb_crud_with_go/internal/httpserver"
	"03-dynamodb_crud_with_go/internal/repository"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var app apptypes.App

func main() {

	run(os.Stdout)
}

func run(out io.Writer) {
	err := loadEnv(envPath)

	if err != nil {
		fmt.Fprintf(out, "[envConfig] WARNING: no .env at %s: %v\n", envPath, err)
	}

	// -----------------------------------------------------------------

	client := db.NewClient(out, context.Background())
	app.Repo = repository.NewRepository(client)

	// -----------------------------------------------------------------
	// Create a context that will be cancelled when an interrupt or termination signal is received.

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	// stop() will stop the context from listening for further OS signals
	defer stop()

	// -----------------------------------------------------------------

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := httpserver.NewHttpServer(port, app)
	fmt.Fprintf(out, "HTTP server listening on %s\n", server.HTTPServer.Addr)

	go func() {
		err := server.HTTPServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(out, "server failed: %v\n", err)
		}
	}()

	// Wait for an OS signal before shutting down gracefully.
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = server.Shutdown(shutdownCtx)

	if err != nil {
		fmt.Fprintf(out, "shutdown failed: %v\n", err)

		closeErr := server.HTTPServer.Close()

		if closeErr != nil {
			fmt.Fprintf(out, "forced shutdown failed: %v\n", closeErr)
		}
	}
}
