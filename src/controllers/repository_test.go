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

func (m *mockRepositoryUseCase) SearchRepositories(query, language, perPage, page string) (*models.RepositorySearchResponse, error) {
	args := m.Called(query, language, perPage, page)
	return args.Get(0).(*models.RepositorySearchResponse), args.Error(1)
}

func (m *mockRepositoryUseCase) ValidateQuery(query string) (language string, err error) {
	args := m.Called(query)
	return args.Get(0).(string), args.Error(1)
}

func TestSearchRepositoriesEndpoint(t *testing.T) {
	tests := map[string]struct {
		endpoint       string
		mockCall       func(*mockRepositoryUseCase)
		expectedStatus int
	}{
		"nominal": {
			endpoint: "/repositories/search?q=golang+language:go",
			mockCall: func(m *mockRepositoryUseCase) {
				m.On("ValidateQuery", "golang language:go").Return("go", nil)
				m.On("SearchRepositories", "golang language:go", "go", "100", "1").Return(&models.RepositorySearchResponse{
					TotalCount: 1,
					Items: []models.Repository{
						{FullName: "scalingo/scalingo-test"},
					},
				}, nil)

			},
			expectedStatus: http.StatusOK,
		},
		"usecase error, return error": {
			endpoint: "/repositories/search?q=golang",
			mockCall: func(m *mockRepositoryUseCase) {
				m.On("ValidateQuery", "golang").Return("go", nil)
				m.On("SearchRepositories", "golang", "go", "100", "1").Return(&models.RepositorySearchResponse{}, errors.New("usecase error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		"missing query, return error": {
			endpoint: "/repositories/search",
			mockCall: func(m *mockRepositoryUseCase) {
				m.On("ValidateQuery", "").Return("", errors.New("query empty"))
			},
			expectedStatus: http.StatusBadRequest,
		},
		"invalid per_page, return error": {
			endpoint: "/repositories/search?q=golang+language:go&per_page=abc",
			mockCall: func(m *mockRepositoryUseCase) {
				m.On("ValidateQuery", "golang language:go").Return("go", nil)
			},
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

func TestValidatePagination(t *testing.T) {
	tests := map[string]struct {
		perPage     string
		page        string
		wantPerPage string
		wantPage    string
		wantErr     assert.ErrorAssertionFunc
	}{
		"default values when empty": {
			perPage:     "",
			page:        "",
			wantPerPage: "100",
			wantPage:    "1",
			wantErr:     assert.NoError,
		},
		"valid values": {
			perPage:     "50",
			page:        "2",
			wantPerPage: "50",
			wantPage:    "2",
			wantErr:     assert.NoError,
		},
		"invalid per_page": {
			perPage: "abc",
			page:    "1",
			wantErr: assert.Error,
		},
		"per_page too high": {
			perPage: "101",
			page:    "1",
			wantErr: assert.Error,
		},
		"per_page negative": {
			perPage: "-1",
			page:    "1",
			wantErr: assert.Error,
		},
		"invalid page": {
			perPage: "100",
			page:    "abc",
			wantErr: assert.Error,
		},
		"negative page": {
			perPage: "100",
			page:    "-1",
			wantErr: assert.Error,
		},
		"zero page": {
			perPage: "100",
			page:    "0",
			wantErr: assert.Error,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			perPage := tt.perPage
			page := tt.page

			err := validatePagination(&perPage, &page)
			tt.wantErr(t, err)
		})
	}
}
