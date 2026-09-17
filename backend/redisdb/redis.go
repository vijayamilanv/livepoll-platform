package redisdb

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"strings"

	"livepoll/config"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

// Connect initialises the Redis client and verifies connectivity.
func Connect() {
	var opts *redis.Options

	if strings.HasPrefix(config.App.RedisAddr, "redis://") || strings.HasPrefix(config.App.RedisAddr, "rediss://") {
		var err error
		opts, err = redis.ParseURL(config.App.RedisAddr)
		if err != nil {
			log.Fatalf("[redis] invalid redis URL: %v", err)
		}
	} else {
		opts = &redis.Options{
			Addr:     config.App.RedisAddr,
			Password: config.App.RedisPassword,
			DB:       0,
		}
		if strings.Contains(config.App.RedisAddr, "upstash.io") {
			opts.TLSConfig = &tls.Config{
				MinVersion: tls.VersionTLS12,
			}
		}
	}

	Client = redis.NewClient(opts)

	ctx := context.Background()
	if err := Client.Ping(ctx).Err(); err != nil {
		log.Fatalf("[redis] ping error: %v", err)
	}
	log.Println("[redis] connected to", config.App.RedisAddr)
}

// CounterKey returns the Redis hash key for a poll's vote counters.
// Hash field = optionId, value = count (int).
func CounterKey(pollID string) string {
	return fmt.Sprintf("poll:%s:counts", pollID)
}

// PubSubChannel returns the Redis pub/sub channel name for a poll.
func PubSubChannel(pollID string) string {
	return fmt.Sprintf("poll:%s:events", pollID)
}

// IncrVote atomically increments the counter for an option.
func IncrVote(ctx context.Context, pollID, optionID string) (int64, error) {
	return Client.HIncrBy(ctx, CounterKey(pollID), optionID, 1).Result()
}

// GetCounts returns a map[optionID]count for a poll from Redis.
// Returns nil, nil if the hash does not exist (cold cache).
func GetCounts(ctx context.Context, pollID string) (map[string]string, error) {
	key := CounterKey(pollID)
	exists, err := Client.Exists(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	if exists == 0 {
		return nil, nil // cache miss
	}
	return Client.HGetAll(ctx, key).Result()
}

// SetCounts sets all option counts for a poll (used during cache warm-up).
func SetCounts(ctx context.Context, pollID string, counts map[string]int64) error {
	key := CounterKey(pollID)
	args := make(map[string]interface{}, len(counts))
	for k, v := range counts {
		args[k] = v
	}
	return Client.HSet(ctx, key, args).Err()
}

// Publish broadcasts a JSON payload to all subscribers of a poll channel.
func Publish(ctx context.Context, pollID, payload string) error {
	return Client.Publish(ctx, PubSubChannel(pollID), payload).Err()
}

// Subscribe returns a *redis.PubSub subscription for a poll channel.
func Subscribe(ctx context.Context, pollID string) *redis.PubSub {
	return Client.Subscribe(ctx, PubSubChannel(pollID))
}
