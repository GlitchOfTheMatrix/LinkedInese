package rateLimiter

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"backend/config"
)

// Limiter holds the Redis client and rate limit settings
type Limiter struct {
	client     *redis.Client
	requests   int
	windowSecs int
}

// New creates a new Limiter, connecting to Upstash Redis
func New(cfg *config.Config) (*Limiter, error) {
	opts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis URL: %w", err)
	}

	client := redis.NewClient(opts)

	// Ping Redis to verify the connection works at startup
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &Limiter{
		client:     client,
		requests:   cfg.RateLimitRequests,
		windowSecs: cfg.RateLimitWindowSeconds,
	}, nil
}

// Allow returns true if the IP is within the rate limit, false if exceeded
func (l *Limiter) Allow(ip string) (bool, error) {
	ctx := context.Background()
	now := time.Now().UnixMilli()
	windowStart := now - int64(l.windowSecs)*1000
	key := fmt.Sprintf("ratelimit:%s", ip)

	// Pipeline batches multiple Redis commands into one round trip
	pipe := l.client.Pipeline()

	// Remove timestamps outside the current window
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart))

	// Add the current request timestamp
	pipe.ZAdd(ctx, key, redis.Z{
		Score:  float64(now),
		Member: now,
	})

	// Count requests in the current window
	countCmd := pipe.ZCard(ctx, key)

	// Set the key to expire after the window
	// so Redis doesn't store data forever
	pipe.Expire(ctx, key, time.Duration(l.windowSecs)*time.Second)

	// Execute all commands at once
	if _, err := pipe.Exec(ctx); err != nil {
		return false, fmt.Errorf("redis pipeline failed: %w", err)
	}

	count := countCmd.Val()
	return count <= int64(l.requests), nil
}
