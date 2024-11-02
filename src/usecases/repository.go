package usecases

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Scalingo/sclng-backend-test-v1/src/models"
	"github.com/Scalingo/sclng-backend-test-v1/src/repositories"
)

// RepositoryUseCase is the interface for the repository use case
type RepositoryUseCase interface {
	SearchRepositories(query string) (*models.RepositorySearchResponse, error)
	ValidateQuery(query string) error
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

// SearchRepositories searches repositories
func (ru *repositoryUseCase) SearchRepositories(q string) (*models.RepositorySearchResponse, error) {
	repos, err := ru.gr.SearchRepositories(q)
	if err != nil {
		return nil, err
	}

	return repos, err
}

// ValidateQuery verifies the query and filters inside it
func (ru *repositoryUseCase) ValidateQuery(q string) error {
	if err := verifyQueryLength(q); err != nil {
		return err
	}

	if err := validateFilters(q); err != nil {
		return err
	}

	return nil
}

// ValidatorFunc is used to validates a filter
type ValidatorFunc func(qualifier, value string) error

// validateFilters verifies the filters in the query
func validateFilters(q string) error {
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
	for _, part := range parts {
		qualifier, value, found := strings.Cut(part, ":")
		if !found {
			continue
		}

		validator, exists := validators[qualifier]
		if !exists {
			return fmt.Errorf("unknown qualifier: %s", qualifier)
		}

		if err := validator(qualifier, value); err != nil {
			return err
		}
	}

	return nil
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
	case strings.HasPrefix(value, ">="):
		extractedValue = value[2:]
	case strings.HasPrefix(value, "<="):
		extractedValue = value[2:]
	case strings.HasPrefix(value, ">"):
		extractedValue = value[1:]
	case strings.HasPrefix(value, "<"):
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
