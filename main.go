package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Scalingo/sclng-backend-test-v1/src/controllers"
	"github.com/Scalingo/sclng-backend-test-v1/src/repositories"
	"github.com/Scalingo/sclng-backend-test-v1/src/usecases"
	"github.com/joho/godotenv"
)

func main() {
	mux := initDependencies()

	cfg, err := newConfig()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Server starting on %d", cfg.Port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), mux); err != nil {
		log.Fatal(err)
	}
}

func initDependencies() *http.ServeMux {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Warning: .env file not found")
	}

	mux := http.NewServeMux()

	rg := repositories.NewGitHubRepository()
	ru := usecases.NewRepositoryUseCase(rg)
	rc := controllers.NewRepositoryController(ru)

	mux.HandleFunc("/repos", rc.SearchRepositories)

	return mux
}
