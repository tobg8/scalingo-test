package usecases

import (
	"errors"
	"fmt"
	"strings"
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

func (m *mockGitHubRepository) GetLanguages(repoFullName string) (models.Languages, error) {
	args := m.Called(repoFullName)
	return args.Get(0).(models.Languages), args.Error(1)
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
	const language = "go"
	query := fmt.Sprintf(" language:%s", language)

	tests := map[string]struct {
		language      string
		query         string
		mockCall      func(*mockGitHubRepository)
		wantError     assert.ErrorAssertionFunc
		checkResponse func(*testing.T, *models.RepositorySearchResponse)
	}{
		"nominal": {
			language: language,
			query:    "tetris" + query,
			mockCall: func(m *mockGitHubRepository) {
				response := &models.RepositorySearchResponse{
					TotalCount: 1,
					Items: []models.Repository{
						{FullName: "scalingo/scalingo-test"},
					},
				}

				m.On("SearchRepositories", "tetris"+query).Return(response, nil)
				m.On("GetLanguages", "scalingo/scalingo-test").Return(models.Languages{"go": 10}, nil)
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
		"error fetching languages": {
			language: language,
			query:    "tetris" + query,
			mockCall: func(m *mockGitHubRepository) {
				response := &models.RepositorySearchResponse{
					TotalCount: 1,
					Items: []models.Repository{
						{FullName: "scalingo/scalingo-test"},
					},
				}
				m.On("SearchRepositories", "tetris"+query).Return(response, nil)
				m.On("GetLanguages", "scalingo/scalingo-test").Return(models.Languages{}, errors.New("API error"))
			},
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
			resp, err := ru.SearchRepositories(tt.query, tt.language)

			tt.wantError(t, err)
			tt.checkResponse(t, resp)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestValidateQuery(t *testing.T) {
	const language = "go"
	query := fmt.Sprintf(" language:%s", language)

	tests := map[string]struct {
		language  string
		query     string
		wantError assert.ErrorAssertionFunc
	}{
		"valid simple query": {
			language:  language,
			query:     "tetris" + query,
			wantError: assert.NoError,
		},
		"valid complex query": {
			language:  language,
			query:     "tetris stars:>100" + query,
			wantError: assert.NoError,
		},
		"empty query, return error": {
			query:     "",
			wantError: assert.Error,
		},
		"exceeds length limit, return error": {
			query:     strings.Repeat("a", 257),
			wantError: assert.Error,
		},
		"invalid filter, return error": {
			query:     "tetris unknown:value",
			wantError: assert.Error,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ru := &repositoryUseCase{}
			language, err := ru.ValidateQuery(tt.query)
			tt.wantError(t, err)
			assert.Equal(t, tt.language, language)
		})
	}
}

func TestValidateFilters(t *testing.T) {
	const language = "go"

	query := fmt.Sprintf(" language:%s", language)
	tests := map[string]struct {
		language  string
		query     string
		wantError assert.ErrorAssertionFunc
	}{
		"valid query with keyword": {
			language:  "go",
			query:     "tetris" + query,
			wantError: assert.NoError,
		},
		"valid query with number operator": {
			language:  "go",
			query:     "size:>=10" + query,
			wantError: assert.NoError,
		},
		"valid query with range": {
			language:  "go",
			query:     "stars:10..20" + query,
			wantError: assert.NoError,
		},
		"valid query with date": {
			language:  "go",
			query:     "created:2024-03-21" + query,
			wantError: assert.NoError,
		},
		"valid complex query": {
			language:  "go",
			query:     "tetris stars:>100 created:>2023-01-01" + query,
			wantError: assert.NoError,
		},
		"unknown qualifier, return error": {
			query:     "unknown:value",
			wantError: assert.Error,
		},
		"missing language filter, return error": {
			query:     "tetris stars:>100 created:>2023-01-01",
			wantError: assert.Error,
		},
		"invalid number format, return error": {
			query:     "stars:abc",
			wantError: assert.Error,
		},
		"invalid date format, return error": {
			query:     "created:2024/03/21",
			wantError: assert.Error,
		},
		"invalid range format, return error": {
			query:     "size:10...20",
			wantError: assert.Error,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			language, err := validateFilters(tt.query)
			tt.wantError(t, err)
			assert.Equal(t, tt.language, language)
		})
	}
}

func TestValidateNumberOperator(t *testing.T) {
	tests := map[string]struct {
		qualifier string
		value     string
		wantError assert.ErrorAssertionFunc
	}{
		"valid number": {
			qualifier: "size",
			value:     "10",
			wantError: assert.NoError,
		},
		"valid number with operator": {
			qualifier: "stars",
			value:     ">=100",
			wantError: assert.NoError,
		},
		"valid range": {
			qualifier: "forks",
			value:     "10..20",
			wantError: assert.NoError,
		},
		"empty value, return error": {
			qualifier: "size",
			value:     "",
			wantError: assert.Error,
		},
		"not a number, return error": {
			qualifier: "stars",
			value:     "abc",
			wantError: assert.Error,
		},
		"operator without number, return error": {
			qualifier: "forks",
			value:     ">=",
			wantError: assert.Error,
		},
		"invalid range format, return error": {
			qualifier: "size",
			value:     "10...20",
			wantError: assert.Error,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := validateNumberOperator(tt.qualifier, tt.value)
			tt.wantError(t, err)
		})
	}
}

func TestValidateDateOperator(t *testing.T) {
	tests := map[string]struct {
		qualifier string
		value     string
		wantError assert.ErrorAssertionFunc
	}{
		"valid date": {
			qualifier: "created",
			value:     "2024-03-21",
			wantError: assert.NoError,
		},
		"valid date with operator": {
			qualifier: "pushed",
			value:     ">=2024-01-01",
			wantError: assert.NoError,
		},
		"empty date, return error": {
			qualifier: "created",
			value:     "",
			wantError: assert.Error,
		},
		"invalid format, return error": {
			qualifier: "pushed",
			value:     "2024/03/21",
			wantError: assert.Error,
		},
		"invalid date, return error": {
			qualifier: "created",
			value:     "2024-13-45",
			wantError: assert.Error,
		},
		"not a date, return error": {
			qualifier: "pushed",
			value:     "hello",
			wantError: assert.Error,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := validateDateOperator(tt.qualifier, tt.value)
			tt.wantError(t, err)
		})
	}
}

func TestExtractValue(t *testing.T) {
	tests := map[string]struct {
		value    string
		expected string
	}{
		"greater than or equal": {
			value:    ">=10",
			expected: "10",
		},
		"less than or equal": {
			value:    "<=20",
			expected: "20",
		},
		"greater than": {
			value:    ">5",
			expected: "5",
		},
		"less than": {
			value:    "<15",
			expected: "15",
		},
		"no operator": {
			value:    "25",
			expected: "25",
		},
		"empty string": {
			value:    "",
			expected: "",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := extractValue(tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateRange(t *testing.T) {
	tests := map[string]struct {
		qualifier string
		value     string
		wantError assert.ErrorAssertionFunc
	}{
		"valid range": {
			qualifier: "size",
			value:     "10..20",
			wantError: assert.NoError,
		},
		"invalid format, return error": {
			qualifier: "size",
			value:     "10...20",
			wantError: assert.Error,
		},
		"not a number, return error": {
			qualifier: "size",
			value:     "sca..lingo",
			wantError: assert.Error,
		},
		"start greater than end, return error": {
			qualifier: "size",
			value:     "20..10",
			wantError: assert.Error,
		},
		"start equals end, return error": {
			qualifier: "size",
			value:     "10..10",
			wantError: assert.Error,
		},
		"single number, return error": {
			qualifier: "size",
			value:     "10",
			wantError: assert.Error,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := validateRange(tt.qualifier, tt.value)
			tt.wantError(t, err)
		})
	}
}

func TestValidateEqualOperator(t *testing.T) {
	tests := map[string]struct {
		qualifier string
		value     string
		wantError assert.ErrorAssertionFunc
	}{
		"valid string value": {
			qualifier: "language",
			value:     "golang",
			wantError: assert.NoError,
		},
		"empty value, return error": {
			qualifier: "language",
			value:     "",
			wantError: assert.Error,
		},
		"numeric value, return error": {
			qualifier: "license",
			value:     "123",
			wantError: assert.Error,
		},
		"only whitespace, return error": {
			qualifier: "language",
			value:     "   ",
			wantError: assert.Error,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := validateEqualOperator(tt.qualifier, tt.value)
			tt.wantError(t, err)
		})
	}
}

func TestVerifyQueryLength(t *testing.T) {
	tests := map[string]struct {
		query     string
		wantError assert.ErrorAssertionFunc
	}{
		"valid length": {
			query:     "golang",
			wantError: assert.NoError,
		},
		"exactly 256 characters": {
			query:     strings.Repeat("a", 256),
			wantError: assert.NoError,
		},
		"empty query, return error": {
			query:     "",
			wantError: assert.Error,
		},
		"exceeds 256 characters, return error": {
			query:     strings.Repeat("a", 257),
			wantError: assert.Error,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := verifyQueryLength(tt.query)
			tt.wantError(t, err)
		})
	}
}
