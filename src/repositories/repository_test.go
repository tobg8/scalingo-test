package repositories

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Scalingo/sclng-backend-test-v1/src/models"
	"github.com/stretchr/testify/assert"
)

func TestNewGitHubRepository(t *testing.T) {
	repo := NewGitHubRepository()
	assert.NotNil(t, repo)

	gr, ok := repo.(*githubRepository)
	assert.True(t, ok)
	assert.Equal(t, "https://api.github.com", gr.baseURL)
	assert.NotNil(t, gr.httpClient)
}

type testCase struct {
	endpoint       string
	rsp            *models.RepositorySearchParams
	mockResponse   string
	mockStatusCode int
	mockServerFunc func(*testing.T, testCase, http.ResponseWriter, *http.Request)
	wantError      assert.ErrorAssertionFunc
}

func setupTestServer(t *testing.T, tt testCase) (*httptest.Server, *githubRepository) {
	if tt.mockServerFunc != nil {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tt.mockServerFunc(t, tt, w, r)
		}))
		return server, &githubRepository{
			baseURL:    server.URL,
			httpClient: server.Client(),
		}
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/vnd.github+json", r.Header.Get("Accept"))
		assert.Equal(t, "2022-11-28", r.Header.Get("X-GitHub-Api-Version"))
		assert.Equal(t, tt.endpoint, r.URL.Path)

		w.WriteHeader(tt.mockStatusCode)
		fmt.Fprintln(w, tt.mockResponse)
	}))

	return server, &githubRepository{
		baseURL:    server.URL,
		httpClient: server.Client(),
	}
}

func TestSearchRepositories(t *testing.T) {
	tests := map[string]testCase{
		"nominal": {
			endpoint: "/search/repositories",
			rsp: &models.RepositorySearchParams{
				Query: "golang",
			},
			mockResponse: `{
				"total_count": 1,
				"incomplete_results": false,
				"items": [{
					"id": 22,
					"full_name": "scalingo/scalingo-test",
					"description": "",
					"html_url": "https://github.com/octocat/Hello-World",
					"owner": {
						"login": "octocat"
					}
				}]
			}`,
			mockStatusCode: http.StatusOK,
			mockServerFunc: func(t *testing.T, tc testCase, w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "application/vnd.github+json", r.Header.Get("Accept"))
				assert.Equal(t, "2022-11-28", r.Header.Get("X-GitHub-Api-Version"))
				assert.Equal(t, tc.endpoint, r.URL.Path)
				assert.Equal(t, tc.rsp.Query, r.URL.Query().Get("q"))
				w.WriteHeader(tc.mockStatusCode)
				fmt.Fprintln(w, tc.mockResponse)
			},
			wantError: assert.NoError,
		},
		"api error": {
			endpoint: "/search/repositories",
			rsp: &models.RepositorySearchParams{
				Query: "golang",
			},
			mockResponse:   `{"message": "API rate limit exceeded"}`,
			mockStatusCode: http.StatusForbidden,
			wantError:      assert.Error,
		},
		"empty query": {
			endpoint: "/search/repositories",
			rsp: &models.RepositorySearchParams{
				Query: "",
			},
			mockResponse:   `{"message": "Validation Failed"}`,
			mockStatusCode: http.StatusUnprocessableEntity,
			wantError:      assert.Error,
		},
		"invalid json response": {
			endpoint: "/search/repositories",
			rsp: &models.RepositorySearchParams{
				Query: "golang",
			},
			mockResponse:   `{invalid json}`,
			mockStatusCode: http.StatusOK,
			wantError:      assert.Error,
		},
		"network error": {
			endpoint: "/search/repositories",
			rsp: &models.RepositorySearchParams{
				Query: "golang",
			},
			wantError: assert.Error,
			mockServerFunc: func(t *testing.T, tc testCase, w http.ResponseWriter, r *http.Request) {
				panic("network error")
			},
		},
		"invalid url": {
			endpoint: "/search/repositories",
			rsp: &models.RepositorySearchParams{
				Query: "golang",
			},
			wantError: assert.Error,
			mockServerFunc: func(t *testing.T, tc testCase, w http.ResponseWriter, r *http.Request) {
				repo := &githubRepository{
					baseURL:    "://invalid-url",
					httpClient: &http.Client{},
				}
				_, err := repo.SearchRepositories(tc.rsp)
				assert.Error(t, err)
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if name == "invalid url" {
				repo := &githubRepository{
					baseURL:    "://invalid-url",
					httpClient: &http.Client{},
				}
				_, err := repo.SearchRepositories(tt.rsp)
				tt.wantError(t, err)
				return
			}

			server, repo := setupTestServer(t, tt)
			defer server.Close()

			result, err := repo.SearchRepositories(tt.rsp)
			tt.wantError(t, err)

			if tt.mockStatusCode == http.StatusOK && err == nil {
				assert.NotNil(t, result)
				var expected models.RepositorySearchResponse
				assert.NoError(t, json.Unmarshal([]byte(tt.mockResponse), &expected))
				assert.Equal(t, expected.TotalCount, result.TotalCount)
				assert.Equal(t, expected.IncompleteResults, result.IncompleteResults)
				assert.Equal(t, len(expected.Items), len(result.Items))

				if len(result.Items) > 0 {
					assert.Equal(t, expected.Items[0].FullName, result.Items[0].FullName)
					assert.Equal(t, expected.Items[0].Description, result.Items[0].Description)
					assert.Equal(t, expected.Items[0].Owner.Login, result.Items[0].Owner.Login)
				}
			}
		})
	}
}

func TestGetLanguages(t *testing.T) {
	tests := map[string]testCase{
		"nominal": {
			endpoint: "/repos/scalingo/scalingo-test/languages",
			rsp: &models.RepositorySearchParams{
				Query: "scalingo/scalingo-test",
			},
			mockResponse: `{
				"Go": 123456,
				"JavaScript": 89012,
				"Python": 45678
			}`,
			mockStatusCode: http.StatusOK,
			wantError:      assert.NoError,
		},
		"repository not found": {
			endpoint: "/repos/scalingo/not-exists/languages",
			rsp: &models.RepositorySearchParams{
				Query: "scalingo/not-exists",
			},
			mockResponse:   `{"message": "Not Found"}`,
			mockStatusCode: http.StatusNotFound,
			wantError:      assert.Error,
		},
		"api error": {
			endpoint: "/repos/scalingo/scalingo-test/languages",
			rsp: &models.RepositorySearchParams{
				Query: "scalingo/scalingo-test",
			},
			mockResponse:   `{"message": "API rate limit exceeded"}`,
			mockStatusCode: http.StatusForbidden,
			wantError:      assert.Error,
		},
		"invalid json response": {
			endpoint: "/repos/scalingo/scalingo-test/languages",
			rsp: &models.RepositorySearchParams{
				Query: "scalingo/scalingo-test",
			},
			mockResponse:   `{invalid json}`,
			mockStatusCode: http.StatusOK,
			wantError:      assert.Error,
		},
		"network error": {
			endpoint: "/repos/scalingo/scalingo-test/languages",
			rsp: &models.RepositorySearchParams{
				Query: "scalingo/scalingo-test",
			},
			wantError: assert.Error,
			mockServerFunc: func(t *testing.T, tc testCase, w http.ResponseWriter, r *http.Request) {
				panic("network error")
			},
		},
		"invalid url": {
			endpoint: "/repos/scalingo/scalingo-test/languages",
			rsp: &models.RepositorySearchParams{
				Query: "scalingo/scalingo-test",
			},
			wantError: assert.Error,
			mockServerFunc: func(t *testing.T, tc testCase, w http.ResponseWriter, r *http.Request) {
				repo := &githubRepository{
					baseURL:    "://invalid-url",
					httpClient: &http.Client{},
				}
				_, err := repo.GetLanguages(tc.rsp.Query, "")
				assert.Error(t, err)
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if name == "invalid url" {
				repo := &githubRepository{
					baseURL:    "://invalid-url",
					httpClient: &http.Client{},
				}
				_, err := repo.GetLanguages(tt.rsp.Query, "")
				tt.wantError(t, err)
				return
			}

			server, repo := setupTestServer(t, tt)
			defer server.Close()

			languages, err := repo.GetLanguages(tt.rsp.Query, "")
			tt.wantError(t, err)

			if tt.mockStatusCode == http.StatusOK && err == nil {
				assert.NotNil(t, languages)
				var expected models.Languages
				assert.NoError(t, json.Unmarshal([]byte(tt.mockResponse), &expected))
				assert.Equal(t, expected, languages)
			}
		})
	}
}
