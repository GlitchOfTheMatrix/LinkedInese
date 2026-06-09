package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"backend/config"
	"backend/rateLimiter"
	"backend/service"
)

type TranslateRequest struct {
	Text string `json:"text"`
	Mode string `json:"mode"`
}

type TranslateResponse struct {
	Result string `json:"result"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// getIP extracts the real client IP from the request
func getIP(r *http.Request) string {
	// When behind a proxy or load balancer, the real IP
	// is in this header, not r.RemoteAddr
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	return r.RemoteAddr
}

func NewTranslateHandler(cfg *config.Config, limiter *rateLimiter.Limiter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Check rate limit before doing anything else
		ip := getIP(r)
		allowed, err := limiter.Allow(ip)
		if err != nil {
			// Redis is down — fail open (allow the request)
			// Better to serve than to block everyone
			log.Printf("rate limiter error: %v", err)
		} else if !allowed {
			writeJSON(w, http.StatusTooManyRequests, ErrorResponse{
				Error: "too many requests, slow down",
			})
			return
		}

		// Decode the request body
		var req TranslateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{
				Error: "invalid request body",
			})
			return
		}

		// Validate inputs
		if req.Text == "" {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{
				Error: "text is required",
			})
			return
		}

		validModes := map[string]bool{
			"translate": true,
			"decode":    true,
			"roast":     true,
		}
		if !validModes[req.Mode] {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{
				Error: "mode must be one of: translate, decode, roast",
			})
			return
		}

		// Call the service layer
		result, err := service.Translate(cfg, req.Text, req.Mode)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{
				Error: "something went wrong",
			})
			return
		}

		writeJSON(w, http.StatusOK, TranslateResponse{Result: result})
	}
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
