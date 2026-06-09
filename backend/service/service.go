package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"backend/config"
)

type groqMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type groqRequest struct {
	Model    string        `json:"model"`
	Messages []groqMessage `json:"messages"`
}

type groqChoice struct {
	Message groqMessage `json:"message"`
}

type groqResponse struct {
	Choices []groqChoice `json:"choices"`
}

var httpClient = &http.Client{
	Timeout: 30 * time.Second,
}

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

func Translate(ctx context.Context, cfg *config.Config, text, mode string) (string, error) {
	system, user := buildPrompt(text, mode)
	if system == "" {
		return "", fmt.Errorf("unknown mode: %s", mode)
	}

	reqBody := groqRequest{
		Model: "llama-3.1-8b-instant",
		Messages: []groqMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to encode request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		"POST",
		"https://api.groq.com/openai/v1/chat/completions",
		bytes.NewBuffer(bodyBytes),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to call Groq API: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read Groq response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("groq API returned status %d: %s", resp.StatusCode, string(respBytes))
	}

	var groqResp groqResponse
	if err := json.Unmarshal(respBytes, &groqResp); err != nil {
		return "", fmt.Errorf("failed to decode Groq response: %w", err)
	}

	if len(groqResp.Choices) == 0 {
		return "", fmt.Errorf("groq returned no choices")
	}

	return groqResp.Choices[0].Message.Content, nil
}
