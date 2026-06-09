package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	APIKey                 string
	RedisURL               string
	Port                   string
	RateLimitRequests      int
	RateLimitWindowSeconds int
}

func Load() *Config {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		log.Fatal("GROQ_API_KEY is not set")
	}

	redisURL := os.Getenv("UPSTASH_REDIS_URL")
	if redisURL == "" {
		log.Fatal("UPSTASH_REDIS_URL is not set")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	rateLimitReqs, err := strconv.Atoi(os.Getenv("RATE_LIMIT_REQUESTS"))
	if err != nil {
		rateLimitReqs = 10
	}

	rateLimitWindow, err := strconv.Atoi(os.Getenv("RATE_LIMIT_WINDOW_SECONDS"))
	if err != nil {
		rateLimitWindow = 60
	}

	return &Config{
		APIKey:                 apiKey,
		RedisURL:               redisURL,
		Port:                   port,
		RateLimitRequests:      rateLimitReqs,
		RateLimitWindowSeconds: rateLimitWindow,
	}
}
