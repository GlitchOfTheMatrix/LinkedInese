package main

import (
	"log"
	"net/http"

	"github.com/joho/godotenv"

	"backend/config"
	"backend/handlers"
	"backend/rateLimiter"
)

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading env vars from system")
	}

	cfg := config.Load()

	// Create the rate limiter — fails fast if Redis is unreachable
	limiter, err := rateLimiter.New(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /translate", handlers.NewTranslateHandler(cfg, limiter))
	mux.HandleFunc("GET /health", handlers.HealthHandler)

	log.Printf("Server starting on port %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, corsMiddleware(mux)); err != nil {
		log.Fatal(err)
	}
}
