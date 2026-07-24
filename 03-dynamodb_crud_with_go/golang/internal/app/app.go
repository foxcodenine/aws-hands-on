package apptypes

import "03-dynamodb_crud_with_go/internal/repository"

type App struct {
	Repo *repository.Repository
}
