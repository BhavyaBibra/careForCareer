package redisinfra

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache implements llm.Cache backed by Redis.
type Cache struct {
	client *redis.Client
}

func NewCache(client *redis.Client) *Cache {
	return &Cache{client: client}
}

func (c *Cache) Get(ctx context.Context, key string) ([]byte, bool) {
	val, err := c.client.Get(ctx, "llm:"+key).Bytes()
	if err != nil {
		return nil, false
	}
	return val, true
}

func (c *Cache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) {
	_ = c.client.Set(ctx, "llm:"+key, value, ttl).Err()
}

// RateLimiter implements a sliding window rate limiter using Redis sorted sets.
type RateLimiter struct {
	client *redis.Client
}

func NewRateLimiter(client *redis.Client) *RateLimiter {
	return &RateLimiter{client: client}
}

// Allow returns true if the request is within rate limit.
// key: e.g. "ip:1.2.3.4" or "user:uuid"
// limit: max requests per window
// window: time window duration
func (r *RateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	now := time.Now().UnixNano()
	windowNs := window.Nanoseconds()
	redisKey := "ratelimit:" + key

	pipe := r.client.TxPipeline()
	pipe.ZRemRangeByScore(ctx, redisKey, "0", fmt.Sprintf("%d", now-windowNs))
	countCmd := pipe.ZCard(ctx, redisKey)
	pipe.ZAdd(ctx, redisKey, redis.Z{Score: float64(now), Member: now})
	pipe.Expire(ctx, redisKey, window+time.Second)

	if _, err := pipe.Exec(ctx); err != nil {
		return false, err
	}
	return countCmd.Val() < int64(limit), nil
}

// CoachDailyCounter tracks coach messages per user per day using Redis.
type CoachDailyCounter struct {
	client *redis.Client
	limit  int
}

func NewCoachDailyCounter(client *redis.Client, limit int) *CoachDailyCounter {
	return &CoachDailyCounter{client: client, limit: limit}
}

// Increment atomically increments and returns (current_count, allowed, error).
func (c *CoachDailyCounter) Increment(ctx context.Context, userID string) (int, bool, error) {
	key := fmt.Sprintf("coach_daily:%s:%s", userID, todayKey())
	pipe := c.client.TxPipeline()
	incrCmd := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, 25*time.Hour) // slightly over 24h to handle timezone edge cases

	if _, err := pipe.Exec(ctx); err != nil {
		return 0, false, err
	}
	count := int(incrCmd.Val())
	return count, count <= c.limit, nil
}

// GetCount returns the current daily message count for a user.
func (c *CoachDailyCounter) GetCount(ctx context.Context, userID string) (int, error) {
	key := fmt.Sprintf("coach_daily:%s:%s", userID, todayKey())
	val, err := c.client.Get(ctx, key).Int()
	if err == redis.Nil {
		return 0, nil
	}
	return val, err
}

func todayKey() string {
	return time.Now().UTC().Format("2006-01-02")
}
