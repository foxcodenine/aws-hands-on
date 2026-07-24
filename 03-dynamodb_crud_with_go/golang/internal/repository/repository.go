package repository

type Repository struct {
	User *UserRepository
}

func NewRepository(client DynamoDBAPI) *Repository {
	return &Repository{
		User: NewUserRepository(client),
	}
}
