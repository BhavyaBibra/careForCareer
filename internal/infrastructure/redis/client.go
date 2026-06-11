package redisinfra

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"

	"github.com/redis/go-redis/v9"
)

func NewClient(addr, password string) (*redis.Client, error) {
	opts := &redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	}
	// Upstash and other hosted Redis providers require TLS. Enable it for
	// any non-local address so the binary works on Render without extra config.
	if !strings.HasPrefix(addr, "localhost") && !strings.HasPrefix(addr, "127.0.0.1") {
		opts.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}
	client := redis.NewClient(opts)
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("redis: connection failed: %w", err)
	}
	return client, nil
}
