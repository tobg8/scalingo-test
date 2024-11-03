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

func (m *mockRepositoryUseCase) SearchRepositories(rsp *models.RepositorySearchParams) (*models.RepositorySearchResponse, error) {
	args := m.Called(rsp)
	return args.Get(0).(*models.RepositorySearchResponse), args.Error(1)
}

func (m *mockRepositoryUseCase) ValidateQuery(query string) (language string, err error) {
	args := m.Called(query)
	return args.Get(0).(string), args.Error(1)
}

func TestSearchRepositoriesEndpoint(t *testing.T) {
	header := "Bearer tokentoken"
	tests := map[string]struct {
		rsp            *models.RepositorySearchParams
		mockCall       func(*mockRepositoryUseCase)
		expectedStatus int
	}{
		"nominal": {
			rsp: &models.RepositorySearchParams{
				Query:    "golang+language:go",
				Header:   header,
				PerPage:  "100",
				Page:     "1",
				Language: "go",
			},
			mockCall: func(m *mockRepositoryUseCase) {
				m.On("ValidateQuery", "golang language:go").Return("go", nil)
				m.On("SearchRepositories", &models.RepositorySearchParams{
					Query:    "golang language:go",
					Header:   header,
					Language: "go",
					PerPage:  "100",
					Page:     "1",
				}).Return(&models.RepositorySearchResponse{
					TotalCount: 1,
					Items: []models.Repository{
						{FullName: "scalingo/scalingo-test"},
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		"missing header token, return error": {
			rsp: &models.RepositorySearchParams{
				Query: "golang",
			},
			mockCall: func(m *mockRepositoryUseCase) {
				m.On("ValidateQuery", "golang").Return("go", nil)
				m.On("SearchRepositories", &models.RepositorySearchParams{
					Query:    "golang",
					Header:   "",
					Language: "",
					PerPage:  "100",
					Page:     "1",
				}).Return(&models.RepositorySearchResponse{}, errors.New("usecase error"))
			},
			expectedStatus: http.StatusUnauthorized,
		},
		"usecase error, return error": {
			rsp: &models.RepositorySearchParams{
				Query:   "wow",
				Header:  header,
				PerPage: "100",
				Page:    "1",
			},
			mockCall: func(m *mockRepositoryUseCase) {
				m.On("ValidateQuery", "wow").Return("", nil)
				m.On("SearchRepositories", &models.RepositorySearchParams{
					Query:    "wow",
					Header:   header,
					Language: "",
					PerPage:  "100",
					Page:     "1",
				}).Return(&models.RepositorySearchResponse{}, errors.New("usecase error"))
			},
			expectedStatus: http.StatusBadRequest,
		},
		"missing query, return error": {
			rsp: &models.RepositorySearchParams{
				Header: header,
			},
			mockCall: func(m *mockRepositoryUseCase) {
				m.On("ValidateQuery", "").Return("", errors.New("query empty"))
			},
			expectedStatus: http.StatusBadRequest,
		},
		"invalid per_page, return error": {
			rsp: &models.RepositorySearchParams{
				Query:  "golang+language:go&per_page=abc",
				Header: header,
			},
			mockCall: func(m *mockRepositoryUseCase) {
				m.On("ValidateQuery", "golang language:go").Return("go", nil)
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			queryURL := ""
			if tt.rsp.Query != "" {
				queryURL = "q=" + tt.rsp.Query
			}

			req := httptest.NewRequest(http.MethodGet, "/repositories/search?"+queryURL, nil)
			w := httptest.NewRecorder()

			mockUseCase := new(mockRepositoryUseCase)
			controller := NewRepositoryController(mockUseCase)

			if tt.mockCall != nil {
				tt.mockCall(mockUseCase)
			}
			if tt.rsp.Header != "" {
				req.Header.Add("Authorization", tt.rsp.Header)
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

// this should be a helper function but we use it only here for now
func ptr(s string) *string {
	return &s
}

func TestValidateHeader(t *testing.T) {
	tests := map[string]struct {
		header  *string
		wantErr assert.ErrorAssertionFunc
	}{
		"valid header": {
			header:  ptr("Bearer validtoken"),
			wantErr: assert.NoError,
		},
		"missing Bearer prefix": {
			header:  ptr("token123"),
			wantErr: assert.Error,
		},
		"empty token": {
			header:  ptr("Bearer "),
			wantErr: assert.Error,
		},
		"empty header": {
			header:  ptr(""),
			wantErr: assert.Error,
		},
		"too short header": {
			header:  ptr("Bearer"),
			wantErr: assert.Error,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := validateHeader(tt.header)
			tt.wantErr(t, err)
		})
	}
}
