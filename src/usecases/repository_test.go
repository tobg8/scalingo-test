package usecases

import (
	"errors"
	"testing"

	"github.com/Scalingo/sclng-backend-test-v1/src/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockGitHubRepository struct {
	mock.Mock
}

func (m *mockGitHubRepository) SearchRepositories(q string) (*models.RepositorySearchResponse, error) {
	args := m.Called(q)
	return args.Get(0).(*models.RepositorySearchResponse), args.Error(1)
}

func TestNewRepositoryUseCase(t *testing.T) {
	mockRepo := &mockGitHubRepository{}
	usecase := NewRepositoryUseCase(mockRepo)

	assert.NotNil(t, usecase)

	ru, ok := usecase.(*repositoryUseCase)
	assert.True(t, ok)
	assert.Equal(t, mockRepo, ru.gr)
}

func TestSearchRepositories(t *testing.T) {
	tests := map[string]struct {
		query         string
		mockCall      func(*mockGitHubRepository)
		wantError     assert.ErrorAssertionFunc
		checkResponse func(*testing.T, *models.RepositorySearchResponse)
	}{
		"nominal": {
			query: "golang",
			mockCall: func(m *mockGitHubRepository) {
				response := &models.RepositorySearchResponse{
					TotalCount: 1,
					Items: []models.Repository{
						{FullName: "scalingo/scalingo-test"},
					},
				}

				m.On("SearchRepositories", "golang").Return(response, nil)
			},
			wantError: assert.NoError,
			checkResponse: func(t *testing.T, resp *models.RepositorySearchResponse) {
				assert.NotNil(t, resp)
				assert.Equal(t, 1, resp.TotalCount)
				assert.Len(t, resp.Items, 1)
				assert.Equal(t, "scalingo/scalingo-test", resp.Items[0].FullName)
			},
		},
		"error search": {
			query: "golang",
			mockCall: func(m *mockGitHubRepository) {
				m.On("SearchRepositories", "golang").Return(&models.RepositorySearchResponse{}, errors.New("could not perform search query"))
			},
			wantError: assert.Error,
			checkResponse: func(t *testing.T, resp *models.RepositorySearchResponse) {
				assert.Nil(t, resp)
			},
		},
		"empty query": {
			query:     "",
			wantError: assert.Error,
			checkResponse: func(t *testing.T, resp *models.RepositorySearchResponse) {
				assert.Nil(t, resp)
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mockRepo := new(mockGitHubRepository)
			if tt.mockCall != nil {
				tt.mockCall(mockRepo)
			}

			ru := NewRepositoryUseCase(mockRepo)
			resp, err := ru.SearchRepositories(tt.query)

			tt.wantError(t, err)
			tt.checkResponse(t, resp)
			mockRepo.AssertExpectations(t)
		})
	}
}
