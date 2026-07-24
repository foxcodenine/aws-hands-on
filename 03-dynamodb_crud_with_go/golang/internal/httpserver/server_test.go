package httpserver

import (
	apptypes "03-dynamodb_crud_with_go/internal/app"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewHTTPServerMountsUserRoutes(t *testing.T) {
	server := NewHttpServer("8080", apptypes.App{})
	req := httptest.NewRequest(http.MethodPost, "/users/", strings.NewReader("{"))
	recorder := httptest.NewRecorder()

	server.HTTPServer.Handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	if server.HTTPServer.Addr != ":8080" {
		t.Fatalf("address = %q, want %q", server.HTTPServer.Addr, ":8080")
	}
}

func TestServerShutdown(t *testing.T) {
	server := NewHttpServer("8080", apptypes.App{})

	if err := server.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown returned error: %v", err)
	}
}
