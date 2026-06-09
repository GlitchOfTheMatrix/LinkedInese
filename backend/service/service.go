// service/translator.go
package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"backend/config"
)

// --- Request structs (what we send to Groq) ---

type groqMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type groqRequest struct {
	Model    string        `json:"model"`
	Messages []groqMessage `json:"messages"`
}

// --- Response structs (what Groq sends back) ---

type groqChoice struct {
	Message groqMessage `json:"message"`
}

type groqResponse struct {
	Choices []groqChoice `json:"choices"`
}

// buildPrompt returns the system and user prompts based on mode
func buildPrompt(text, mode string) (system string, user string) {
	switch mode {
	case "translate":
		system = "You are a LinkedIn content expert. Convert plain text into corporate LinkedIn buzzword-filled posts. Use phrases like 'synergy', 'value-add', 'thought leader', 'pivot', 'ecosystem'. Make it sound impressive but hollow."
		user = fmt.Sprintf("Convert this to LinkedInese: %s", text)
	case "decode":
		system = "You decode LinkedIn corporate speak into plain, honest English. Strip away all buzzwords and reveal what is actually being said."
		user = fmt.Sprintf("Decode this LinkedIn post into plain English: %s", text)
	case "roast":
		system = "You are a witty critic who roasts LinkedIn posts. Be funny, sharp, and brutally honest about the corporate hollow-speak."
		user = fmt.Sprintf("Roast this LinkedIn post: %s", text)
	default:
		return "", ""
	}
	return system, user
}

// Translate calls Groq API and returns the translated text
func Translate(cfg *config.Config, text, mode string) (string, error) {
	// Build prompts based on mode
	system, user := buildPrompt(text, mode)
	if system == "" {
		return "", fmt.Errorf("unknown mode: %s", mode)
	}

	// Build the request body
	reqBody := groqRequest{
		Model: "llama-3.1-8b-instant",
		Messages: []groqMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
	}

	// Encode the struct to JSON bytes
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to encode request: %w", err)
	}

	// Create the HTTP request
	httpReq, err := http.NewRequest(
		"POST",
		"https://api.groq.com/openai/v1/chat/completions",
		bytes.NewBuffer(bodyBytes),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	httpReq.Header.Set("Content-Type", "application/json")

	// Make the HTTP call
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to call Groq API: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("groq API returned status: %d", resp.StatusCode)
	}

	// Read and decode the response
	var groqResp groqResponse
	if err := json.NewDecoder(resp.Body).Decode(&groqResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract the text from the first choice
	if len(groqResp.Choices) == 0 {
		return "", fmt.Errorf("groq returned no choices")
	}

	return groqResp.Choices[0].Message.Content, nil
}
