package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Scalingo/sclng-backend-test-v1/src/usecases"
)

type RepositoryController struct {
	ru usecases.RepositoryUseCase
}

func NewRepositoryController(ru usecases.RepositoryUseCase) *RepositoryController {
	return &RepositoryController{
		ru: ru,
	}
}

func (rc *RepositoryController) SearchRepositories(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")

	language, err := rc.ru.ValidateQuery(query)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	perPage := r.URL.Query().Get("per_page")
	page := r.URL.Query().Get("page")

	err = validatePagination(&perPage, &page)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	repos, err := rc.ru.SearchRepositories(query, language, perPage, page)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(repos)
}

func validatePagination(perPage, page *string) error {
	if *perPage == "" {
		*perPage = "100"
	}

	pp, err := strconv.Atoi(*perPage)
	if err != nil || pp < 0 || pp > 100 {
		return fmt.Errorf("per_page must be a number between 0 and 100")
	}

	if *page == "" {
		*page = "1"
	}

	p, err := strconv.Atoi(*page)
	if err != nil || p < 1 {
		return fmt.Errorf("page must be a positive number")
	}

	return nil
}
