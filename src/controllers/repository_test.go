package controllers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Scalingo/sclng-backend-test-v1/src/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockRepositoryUseCase struct {
	mock.Mock
}

func (m *mockRepositoryUseCase) SearchRepositories(query string) (*models.RepositorySearchResponse, error) {
	args := m.Called(query)
	return args.Get(0).(*models.RepositorySearchResponse), args.Error(1)
}

func TestSearchRepositoriesEndpoint(t *testing.T) {
	tests := map[string]struct {
		endpoint       string
		mockCall       func(*mockRepositoryUseCase)
		expectedStatus int
	}{
		"nominal": {
			endpoint: "/repositories/search?q=golang",
			mockCall: func(m *mockRepositoryUseCase) {
				response := &models.RepositorySearchResponse{
					TotalCount: 1,
					Items: []models.Repository{
						{FullName: "scalingo/scalingo-test"},
					},
				}
				m.On("SearchRepositories", "golang").Return(response, nil)
			},
			expectedStatus: http.StatusOK,
		},
		"usecase error, return error": {
			endpoint: "/repositories/search?q=golang",
			mockCall: func(m *mockRepositoryUseCase) {
				m.On("SearchRepositories", "golang").Return(&models.RepositorySearchResponse{}, errors.New("usecase error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		"missing query, return error": {
			endpoint:       "/repositories/search",
			expectedStatus: http.StatusBadRequest,
		},
		"empty queryu, return error": {
			endpoint:       "/repositories/search?q=",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.endpoint, nil)
			w := httptest.NewRecorder()

			mockUseCase := new(mockRepositoryUseCase)
			controller := NewRepositoryController(mockUseCase)

			if tt.mockCall != nil {
				tt.mockCall(mockUseCase)
			}

			controller.SearchRepositories(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}