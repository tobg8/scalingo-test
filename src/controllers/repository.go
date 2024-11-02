package controllers

import (
	"encoding/json"
	"net/http"

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

	repos, err := rc.ru.SearchRepositories(query, language)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(repos)
}
