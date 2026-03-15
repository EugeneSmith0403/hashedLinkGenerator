package limiter

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

var tokenBucketScript = redis.NewScript(`
local key        = KEYS[1]
local capacity   = tonumber(ARGV[1])
local rate       = tonumber(ARGV[2])
local now_ms     = tonumber(ARGV[3])
local cost       = tonumber(ARGV[4])

local data = redis.call("HMGET", key, "tokens", "last_ms")
local tokens  = tonumber(data[1])
local last_ms = tonumber(data[2])

if tokens == nil then
    tokens  = capacity
    last_ms = now_ms
end

local elapsed_sec = (now_ms - last_ms) / 1000
tokens = math.min(capacity, tokens + elapsed_sec * rate)
last_ms = now_ms

local allowed = 0
if tokens >= cost then
    tokens  = tokens - cost
    allowed = 1
end

local ttl = math.ceil(capacity / rate) + 1
redis.call("HMSET", key, "tokens", tokens, "last_ms", last_ms)
redis.call("EXPIRE", key, ttl)

return { allowed, tostring(tokens) }
`)

type KeyType string

const (
	KeyByAccountID KeyType = "account"
	KeyByIP        KeyType = "ip"
)

type Config struct {
	// Capacity is the maximum number of tokens in the bucket (burst limit).
	Capacity int
	// RefillRate is the number of tokens added per second.
	RefillRate float64
	// KeyType defines the key strategy: by accountId (KeyByAccountID) or by IP (KeyByIP).
	KeyType KeyType
	// Cost is the number of tokens consumed per request. Defaults to 1.
	Cost int
}

type Result struct {
	Allowed   bool
	Remaining float64
}

type LimiterService struct {
	rdb    *redis.Client
	config Config
}

func NewLimiter(rdb *redis.Client, cfg Config) *LimiterService {
	if cfg.Cost == 0 {
		cfg.Cost = 1
	}
	return &LimiterService{rdb: rdb, config: cfg}
}

func (s *LimiterService) Allow(ctx context.Context, value string) (Result, error) {
	key := fmt.Sprintf("rl:%s:%s", s.config.KeyType, value)
	nowMs := time.Now().UnixMilli()

	res, err := tokenBucketScript.Run(ctx, s.rdb,
		[]string{key},
		s.config.Capacity,
		s.config.RefillRate,
		nowMs,
		s.config.Cost,
	).Slice()
	if err != nil {
		return Result{}, fmt.Errorf("limiter: redis script error: %w", err)
	}

	allowed := res[0].(int64) == 1
	remaining := parseFloat(res[1])

	return Result{Allowed: allowed, Remaining: remaining}, nil
}

func parseFloat(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case string:
		var f float64
		fmt.Sscanf(val, "%f", &f)
		return f
	}
	return 0
}
