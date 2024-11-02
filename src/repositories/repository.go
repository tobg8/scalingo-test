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
	GetLanguages(repoFullName string) (models.Languages, error)
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

// doRequest is a helper function that handles HTTP request
func (gr *githubRepository) doRequest(endpoint string, result interface{}) error {
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	// Add GitHub API headers
	req.Header.Add("Accept", "application/vnd.github+json")
	req.Header.Add("X-GitHub-Api-Version", "2022-11-28")

	resp, err := gr.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error with request, status: %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("error decoding response: %w", err)
	}

	return nil
}

func (gr *githubRepository) SearchRepositories(query string) (*models.RepositorySearchResponse, error) {
	endpoint := fmt.Sprintf("%s/search/repositories?q=%s", gr.baseURL, url.QueryEscape(query))

	var result models.RepositorySearchResponse
	if err := gr.doRequest(endpoint, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (gr *githubRepository) GetLanguages(repoFullName string) (models.Languages, error) {
	endpoint := fmt.Sprintf("%s/repos/%s/languages", gr.baseURL, repoFullName)

	languages := make(models.Languages)
	if err := gr.doRequest(endpoint, &languages); err != nil {
		return nil, err
	}

	return languages, nil
}
