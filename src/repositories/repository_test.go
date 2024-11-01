package repositories

import (
	"net/http"
	"net/http/httptest"
	"testing"

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

func TestSearchRepositories(t *testing.T) {
	tests := map[string]struct {
		name           string
		query          string
		mockResponse   string
		mockStatusCode int
		wantError      assert.ErrorAssertionFunc
	}{
		"nominal": {
			query: "golang",
			mockResponse: `{
				"total_count": 1,
				"incomplete_results": false,
				"items": [{
					"id": 22,
					"node_id": "azezaeeza",
					"name": "scalingo-test",
					"full_name": "scalingo/scalingo-test",
					"description": "",
					"html_url": "https://github.com/octocat/Hello-World",
					"owner": {
						"login": "octocat",
						"id": 1,
						"node_id": "MDQ6VXNlcjE=",
						"avatar_url": "https://github.com/images/error/octocat_happy.gif",
						"html_url": "https://github.com/octocat"
					}
				}]
			}`,
			mockStatusCode: http.StatusOK,
			wantError:      assert.NoError,
		},
		"api error": {
			query:          "golang",
			mockResponse:   `{"message": "API rate limit exceeded"}`,
			mockStatusCode: http.StatusForbidden,
			wantError:      assert.Error,
		},
		"empty query": {
			query:          "",
			mockResponse:   `{"message": "Validation Failed"}`,
			mockStatusCode: http.StatusUnprocessableEntity,
			wantError:      assert.Error,
		},
	}

	for n, tt := range tests {
		t.Run(n, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "application/vnd.github+json", r.Header.Get("Accept"))
				assert.Equal(t, "2022-11-28", r.Header.Get("X-GitHub-Api-Version"))
				w.WriteHeader(tt.mockStatusCode)
				w.Write([]byte(tt.mockResponse))
			}))
			defer server.Close()

			repo := &githubRepository{
				baseURL:    server.URL,
				httpClient: server.Client(),
			}

			_, err := repo.SearchRepositories(tt.query)
			tt.wantError(t, err)
		})
	}
}
