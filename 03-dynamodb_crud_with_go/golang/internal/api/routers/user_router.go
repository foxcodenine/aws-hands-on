package routers

import (
	"03-dynamodb_crud_with_go/internal/api/handlers"
	apptypes "03-dynamodb_crud_with_go/internal/app"

	"github.com/go-chi/chi/v5"
)

func UserRoutes(app apptypes.App) chi.Router {
	r := chi.NewRouter()

	userHandler := &handlers.UserHandler{App: &app}

	r.Post("/", userHandler.Store)
	r.Get("/", userHandler.Index)
	r.Get("/active", userHandler.Active)
	r.Get("/by-email", userHandler.ByEmail)
	r.Get("/{userID}", userHandler.Show)
	r.Put("/{userID}", userHandler.Update)
	r.Delete("/{userID}", userHandler.Destroy)

	return r
}
