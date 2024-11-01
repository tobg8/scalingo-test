package usecases

import (
	"fmt"

	"github.com/Scalingo/sclng-backend-test-v1/src/models"
	"github.com/Scalingo/sclng-backend-test-v1/src/repositories"
)

type RepositoryUseCase interface {
	SearchRepositories(query string) (*models.RepositorySearchResponse, error)
}

type repositoryUseCase struct {
	gr repositories.GitHubRepository
}

func NewRepositoryUseCase(gr repositories.GitHubRepository) RepositoryUseCase {
	return &repositoryUseCase{
		gr: gr,
	}
}

func (ru *repositoryUseCase) SearchRepositories(q string) (*models.RepositorySearchResponse, error) {
	if q == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	repos, err := ru.gr.SearchRepositories(q)
	if err != nil {
		return nil, err
	}

	return repos, err
}
