package usecases

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Scalingo/sclng-backend-test-v1/src/models"
	"github.com/Scalingo/sclng-backend-test-v1/src/repositories"
)

// RepositoryUseCase is the interface for the repository use case
type RepositoryUseCase interface {
	SearchRepositories(query string, language string) (*models.RepositorySearchResponse, error)
	ValidateQuery(query string) (language string, err error)
}

type repositoryUseCase struct {
	gr repositories.GitHubRepository
}

// NewRepositoryUseCase creates a new repository use case
func NewRepositoryUseCase(gr repositories.GitHubRepository) RepositoryUseCase {
	return &repositoryUseCase{
		gr: gr,
	}
}

// SearchRepositories searches repositories and fetches their languages concurrently
func (ru *repositoryUseCase) SearchRepositories(q string, language string) (*models.RepositorySearchResponse, error) {
	repos, err := ru.gr.SearchRepositories(q)
	if err != nil {
		log.Print("error searching repositories: ", err)
		return nil, err
	}

	errChan := make(chan error, len(repos.Items))
	var wg sync.WaitGroup

	var mu sync.Mutex
	clientRepos := make([]models.Repository, 0, len(repos.Items))

	// For each repository, start a goroutine to fetch its languages
	for i := range repos.Items {
		wg.Add(1)
		repo := repos.Items[i]

		go func() {
			defer wg.Done()

			languages, err := ru.gr.GetLanguages(repo.FullName)
			if err != nil {
				log.Print("error fetching languages for ", repo.FullName, ": ", err)
				errChan <- fmt.Errorf("error fetching languages for %s: %w", repo.FullName, err)
				return
			}

			// Filter languages to only keep the requested language
			filteredLanguages := make(models.Languages)
			queryLanguage := strings.ToUpper(language)

			// Convert and check each language from the repo
			for repoLang, langBytes := range languages {
				if strings.ToUpper(repoLang) == queryLanguage {
					filteredLanguages[repoLang] = langBytes
					break
				}
			}

			// If the repository has the requested language (useless i think it has to but just in case)
			if len(filteredLanguages) > 0 {
				repo.Languages = filteredLanguages
				mu.Lock()
				clientRepos = append(clientRepos, repo)
				mu.Unlock()
			}
		}()
	}

	wg.Wait()
	close(errChan)

	if err := <-errChan; err != nil {
		log.Print("error fetching repository languages: ", err)
		return nil, fmt.Errorf("error fetching repository languages: %w", err)
	}

	return &models.RepositorySearchResponse{
		TotalCount:        repos.TotalCount,
		Count:             len(clientRepos),
		IncompleteResults: repos.IncompleteResults,
		Items:             clientRepos,
	}, nil
}

// ValidateQuery verifies the query and filters inside it
func (ru *repositoryUseCase) ValidateQuery(q string) (language string, err error) {
	if err := verifyQueryLength(q); err != nil {
		return "", err
	}

	if language, err = validateFilters(q); err != nil {
		return "", err
	}

	return language, nil
}

// ValidatorFunc is used to validates a filter
type ValidatorFunc func(qualifier, value string) error

// validateFilters verifies the filters in the query
func validateFilters(q string) (language string, error error) {
	validators := map[string]ValidatorFunc{
		"size":      validateNumberOperator,
		"topics":    validateNumberOperator,
		"stars":     validateNumberOperator,
		"followers": validateNumberOperator,
		"forks":     validateNumberOperator,

		"license":  validateEqualOperator,
		"language": validateEqualOperator,

		"created": validateDateOperator,
		"pushed":  validateDateOperator,
	}

	parts := strings.Fields(q)
	hasLanguageFilter := false

	for _, part := range parts {
		if strings.Count(part, ":") > 1 {
			return "", fmt.Errorf("invalid filter format in '%s': use '+' to separate filters, not ':'", part)
		}

		qualifier, value, found := strings.Cut(part, ":")
		if !found {
			continue
		}

		if qualifier == "language" {
			hasLanguageFilter = true
			language = value
		}

		validator, exists := validators[qualifier]
		if !exists {
			return language, fmt.Errorf("unknown qualifier: %s", qualifier)
		}

		if err := validator(qualifier, value); err != nil {
			return language, err
		}
	}

	if !hasLanguageFilter {
		return language, fmt.Errorf("no language filter set, please provide one")
	}

	return language, nil
}

// validateNumberOperator verifies number filters
func validateNumberOperator(qualifier, value string) error {
	if value == "" {
		return fmt.Errorf("%s cannot be empty", qualifier)
	}

	if strings.Contains(value, "..") {
		err := validateRange(qualifier, value)
		if err != nil {
			return err
		}
		return nil
	}

	number := extractValue(value)
	if number == "" {
		return fmt.Errorf("%s must have a number after the comparison operator", qualifier)
	}

	_, err := strconv.Atoi(number)
	if err != nil {
		return fmt.Errorf("%s must be a number with valid optional comparison operator", qualifier)
	}

	return nil
}

// validateDateOperator verifies date filters
func validateDateOperator(qualifier, value string) error {
	if value == "" {
		return fmt.Errorf("%s cannot be empty", qualifier)
	}

	date := extractValue(value)

	_, err := time.Parse("2006-01-02", date)
	if err != nil {
		return fmt.Errorf("%s must be a valid date in YYYY-MM-DD format, got '%s'", qualifier, value)
	}

	return nil
}

// extractValue extracts the value after the comparison operator if any
func extractValue(value string) string {
	var extractedValue string
	switch {
	case strings.HasPrefix(value, ">=") || strings.HasPrefix(value, "<="):
		extractedValue = value[2:]
	case strings.HasPrefix(value, ">") || strings.HasPrefix(value, "<"):
		extractedValue = value[1:]
	default:
		extractedValue = value
	}
	return extractedValue
}

// validateRange verifies the range format
func validateRange(qualifier, value string) error {
	rangeParts := strings.Split(value, "..")
	if len(rangeParts) != 2 {
		return fmt.Errorf("%s must be a valid range with two numbers separated by '..', got '%s'", qualifier, value)
	}

	start, err1 := strconv.Atoi(rangeParts[0])
	end, err2 := strconv.Atoi(rangeParts[1])
	if err1 != nil || err2 != nil {
		return fmt.Errorf("%s range must contain valid numbers, got '%s'", qualifier, value)
	}

	if start >= end {
		return fmt.Errorf("%s range start must be less than end, got '%s'", qualifier, value)
	}

	return nil
}

// validateEqualOperator verifies all filters with equal operator
// TODO: We naively assume the value is a valid one, language, license...
// TODO: We should fetch the list of possible values from the github API and verify value
func validateEqualOperator(qualifier, value string) error {
	if value == "" {
		return fmt.Errorf("%s cannot be empty", qualifier)
	}

	if _, err := strconv.Atoi(value); err == nil {
		return fmt.Errorf("%s cannot be a number, must be a string", qualifier)
	}

	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s cannot be only whitespace", qualifier)
	}

	return nil
}

// Github prevents a query to be longer than 256 characters
// https://docs.github.com/fr/rest/search/search?apiVersion=2022-11-28#limitations-on-query-length
func verifyQueryLength(query string) error {
	if len(query) == 0 {
		return fmt.Errorf("search query cannot be empty")
	}

	if len(query) > 256 {
		return fmt.Errorf("search query exceeds 256 characters limit")
	}
	return nil
}
