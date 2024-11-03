package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Scalingo/sclng-backend-test-v1/src/models"
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

type ErrorResponse struct {
	Message string `json:"error"`
}

func renderError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{
		Message: message,
	})
}

func (rc *RepositoryController) SearchRepositories(w http.ResponseWriter, r *http.Request) {
	header := r.Header.Get("Authorization")
	err := validateHeader(&header)
	if err != nil {
		renderError(w, http.StatusUnauthorized, err.Error())
		return
	}

	query := r.URL.Query().Get("q")
	language, err := rc.ru.ValidateQuery(query)
	if err != nil {
		renderError(w, http.StatusBadRequest, err.Error())
		return
	}

	perPage := r.URL.Query().Get("per_page")
	page := r.URL.Query().Get("page")

	err = validatePagination(&perPage, &page)
	if err != nil {
		renderError(w, http.StatusBadRequest, err.Error())
		return
	}

	params := models.RepositorySearchParams{
		Query:    query,
		PerPage:  perPage,
		Page:     page,
		Header:   header,
		Language: language,
	}

	repos, err := rc.ru.SearchRepositories(&params)
	if err != nil {
		renderError(w, http.StatusBadRequest, err.Error())
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

func validateHeader(h *string) error {
	if h == nil || *h == "" {
		return fmt.Errorf("missing Authorization header")
	}

	parts := strings.Split(*h, " ")
	if len(parts) != 2 || parts[0] != "Bearer" || parts[1] == "" {
		return fmt.Errorf("invalid Authorization header format, must be 'Bearer token'")
	}

	return nil
}
