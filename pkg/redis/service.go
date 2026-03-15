package redis

import (
	"context"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

type Redis struct {
	client       *redis.Client
	ExpiredCache time.Duration
}

func NewRedis(options *redis.Options, expired time.Duration) *Redis {
	return &Redis{
		client:       redis.NewClient(options),
		ExpiredCache: expired,
	}
}

func (r Redis) Get(key string, defaultValue ...string) string {
	recievedKey, err := r.client.Get(context.Background(), key).Result()

	if err != nil {
		if err != redis.Nil {
			log.Printf("[redis] Get key=%s error: %v", key, err)
		}
		if len(defaultValue) == 0 {
			return ""
		}
		return defaultValue[len(defaultValue)-1]
	}

	return recievedKey
}

func (r Redis) Set(key string, value any, exp time.Duration) string {
	result, err := r.client.Set(context.Background(), key, value, exp).Result()
	if err != nil {
		log.Printf("[redis] Set key=%s error: %v", key, err)
	}
	return result
}

func (r Redis) Incr(key string) int64 {
	result, err := r.client.Incr(context.Background(), key).Result()
	if err != nil {
		log.Printf("[redis] Incr key=%s error: %v", key, err)
	}
	return result
}

func (r Redis) IncrBy(key string, value int64) int64 {
	result, err := r.client.IncrBy(context.Background(), key, value).Result()
	if err != nil {
		log.Printf("[redis] IncrBy key=%s error: %v", key, err)
	}
	return result
}

func (r Redis) Decr(key string) int64 {
	result, err := r.client.Decr(context.Background(), key).Result()
	if err != nil {
		log.Printf("[redis] Decr key=%s error: %v", key, err)
	}
	return result
}

func (r Redis) DecrBy(key string, value int64) int64 {
	result, err := r.client.DecrBy(context.Background(), key, value).Result()
	if err != nil {
		log.Printf("[redis] DecrBy key=%s error: %v", key, err)
	}
	return result
}

func (r Redis) SAdd(key string, members ...any) int64 {
	result, err := r.client.SAdd(context.Background(), key, members...).Result()
	if err != nil {
		log.Printf("[redis] SAdd key=%s error: %v", key, err)
	}
	return result
}

func (r Redis) SMembers(key string) []string {
	result, err := r.client.SMembers(context.Background(), key).Result()
	if err != nil {
		log.Printf("[redis] SMembers key=%s error: %v", key, err)
	}
	return result
}

func (r Redis) SRem(key string, members ...any) int64 {
	result, err := r.client.SRem(context.Background(), key, members...).Result()
	if err != nil {
		log.Printf("[redis] SRem key=%s error: %v", key, err)
	}
	return result
}

func (r Redis) SIsMember(key string, member any) bool {
	result, err := r.client.SIsMember(context.Background(), key, member).Result()
	if err != nil {
		log.Printf("[redis] SIsMember key=%s error: %v", key, err)
	}
	return result
}

func (r Redis) Del(keys ...string) int64 {
	result, err := r.client.Del(context.Background(), keys...).Result()
	if err != nil {
		log.Printf("[redis] Del keys=%v error: %v", keys, err)
	}
	return result
}

func (r Redis) DelByPattern(pattern string) int64 {
	var cursor uint64
	var deleted int64
	for {
		keys, nextCursor, err := r.client.Scan(context.Background(), cursor, pattern, 100).Result()
		if err != nil {
			log.Printf("[redis] DelByPattern pattern=%s scan error: %v", pattern, err)
			break
		}
		if len(keys) > 0 {
			deleted += r.Del(keys...)
		}
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	return deleted
}

func (r Redis) Client() *redis.Client {
	return r.client
}

func (r Redis) Close() {
	r.client.Close()
}
