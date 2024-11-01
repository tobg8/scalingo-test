package repositories

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/Scalingo/sclng-backend-test-v1/src/models"
)

type GitHubRepository interface {
	SearchRepositories(query string) (*models.RepositorySearchResponse, error)
}

type githubRepository struct {
	baseURL    string
	httpClient *http.Client
}

func NewGitHubRepository() GitHubRepository {
	return &githubRepository{
		baseURL:    "https://api.github.com",
		httpClient: &http.Client{},
	}
}

func (r *githubRepository) SearchRepositories(query string) (*models.RepositorySearchResponse, error) {
	endpoint := fmt.Sprintf("%s/search/repositories?q=%s", r.baseURL, url.QueryEscape(query))

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// recommended by the github api documentation
	req.Header.Add("Accept", "application/vnd.github+json")
	req.Header.Add("X-GitHub-Api-Version", "2022-11-28")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result models.RepositorySearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &result, nil
}
