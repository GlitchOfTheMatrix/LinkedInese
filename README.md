# LinkedInese Backend

This is the backend service for **LinkedInese**, a web application that translates plain text into corporate LinkedIn buzzword-filled posts, decodes LinkedIn corporate speak into plain English, and roasts LinkedIn posts.

The backend is built with **Go** and uses the standard `net/http` package for routing. It integrates with **Groq** for LLM capabilities (using the `llama-3.1-8b-instant` model) and uses **Upstash Redis** for rate limiting.

## Features

- **Translate**: Convert plain text into corporate LinkedIn jargon.
- **Decode**: Strip away buzzwords and reveal what a LinkedIn post is actually saying.
- **Roast**: Generate a witty and brutally honest roast of a LinkedIn post.
- **Rate Limiting**: IP-based rate limiting using Redis to prevent abuse.
- **CORS Support**: Pre-configured CORS middleware.

## Prerequisites

- Go 1.21 or higher
- Redis instance (e.g., [Upstash Redis](https://upstash.com/))
- Groq API Key (get it from [Groq Console](https://console.groq.com/))

## Environment Variables

Create a `.env` file in the root of the `backend` directory with the following variables:

```env
# Required
GROQ_API_KEY=your_groq_api_key_here
UPSTASH_REDIS_URL=redis://your_upstash_redis_url:port

# Optional
PORT=8080 # Default is 8080
RATE_LIMIT_REQUESTS=10 # Number of allowed requests per window (Default: 10)
RATE_LIMIT_WINDOW_SECONDS=60 # Rate limit window in seconds (Default: 60)
```

## Running the Server

1. Install dependencies:
   ```bash
   go mod download
   ```

2. Run the application:
   ```bash
   go run main.go
   ```

The server will start on the port specified in your `.env` file (default is `8080`).

## API Endpoints

### 1. Translate / Decode / Roast

**Endpoint:** `POST /translate`

**Request Body:**
```json
{
  "text": "Your text goes here",
  "mode": "translate" // Must be one of: "translate", "decode", "roast"
}
```

**Response:**
```json
{
  "result": "The generated response from Groq."
}
```

**Error Responses:**
- `400 Bad Request`: Missing text, invalid mode, or invalid JSON.
- `429 Too Many Requests`: Rate limit exceeded for the client's IP.
- `500 Internal Server Error`: Something went wrong while calling the Groq API.

### 2. Health Check

**Endpoint:** `GET /health`

**Response:**
```json
{
  "status": "ok"
}
```

## Project Structure

- `main.go`: Application entry point, server setup, and route registration.
- `config/`: Configuration loader for environment variables.
- `handlers/`: HTTP request handlers and validation logic.
- `rateLimiter/`: Redis-based rate limiting logic.
- `service/`: Core business logic and integration with the Groq API.
