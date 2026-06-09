package handlers

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strings"

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

func getIP(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

func NewTranslateHandler(cfg *config.Config, limiter *rateLimiter.Limiter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

		ip := getIP(r)
		allowed, err := limiter.Allow(ip)
		if err != nil {
			log.Printf("rate limiter error for ip %s: %v", ip, err)
		} else if !allowed {
			writeJSON(w, http.StatusTooManyRequests, ErrorResponse{
				Error: "too many requests, slow down",
			})
			return
		}

		var req TranslateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{
				Error: "invalid request body",
			})
			return
		}

		if strings.TrimSpace(req.Text) == "" {
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

		result, err := service.Translate(cfg, req.Text, req.Mode)
		if err != nil {
			log.Printf("translate error: %v", err)
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
