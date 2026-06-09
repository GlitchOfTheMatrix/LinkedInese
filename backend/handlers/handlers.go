package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"strings"

	"backend/config"
	"backend/rateLimiter"
	"backend/service"
)

var validModes = map[string]bool{
	"translate": true,
	"decode":    true,
	"roast":     true,
}

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

func decodeRequest(w http.ResponseWriter, r *http.Request, dst *TranslateRequest) bool {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeJSON(w, http.StatusRequestEntityTooLarge, ErrorResponse{Error: "request body too large"})
			return false
		}
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return false
	}

	if dec.More() {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return false
	}

	return true
}

func validateRequest(w http.ResponseWriter, req TranslateRequest) bool {
	if strings.TrimSpace(req.Text) == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "text is required"})
		return false
	}

	if !validModes[req.Mode] {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "mode must be one of: translate, decode, roast"})
		return false
	}

	return true
}

func checkRateLimit(w http.ResponseWriter, limiter *rateLimiter.Limiter, ip string) bool {
	allowed, err := limiter.Allow(ip)
	if err != nil {
		log.Printf("rate limiter error for ip %s: %v", ip, err)
		return true
	}
	if !allowed {
		writeJSON(w, http.StatusTooManyRequests, ErrorResponse{Error: "too many requests, slow down"})
		return false
	}
	return true
}

func NewTranslateHandler(cfg *config.Config, limiter *rateLimiter.Limiter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

		if !checkRateLimit(w, limiter, getIP(r)) {
			return
		}

		var req TranslateRequest
		if !decodeRequest(w, r, &req) {
			return
		}

		if !validateRequest(w, req) {
			return
		}

		result, err := service.Translate(r.Context(), cfg, req.Text, req.Mode)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			log.Printf("translate error: %v", err)
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "something went wrong"})
			return
		}

		writeJSON(w, http.StatusOK, TranslateResponse{Result: result})
	}
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
