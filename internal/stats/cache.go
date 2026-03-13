package stats

import (
	"link-generator/pkg/redis"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

func statCacheKey(email string, linkID uint, queries map[string]time.Time) (string, error) {
	hashQuery, err := redis.HashFilters(queries)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("email:%s:link:%d:report:%s", email, linkID, hashQuery), nil
}

func GetCachedStat[T any](r *redis.Redis, queries map[string]time.Time, email string, linkID uint) (T, error) {
	var zero T

	key, err := statCacheKey(email, linkID, queries)
	if err != nil {
		return zero, err
	}

	statsData := r.Get(key)
	if statsData == "" {
		return zero, nil
	}

	var result T
	if err := json.Unmarshal([]byte(statsData), &result); err != nil {
		return zero, err
	}

	return result, nil
}

func SetCachedStat[T any](r *redis.Redis, data T, queries map[string]time.Time, email string, linkID uint) {
	key, err := statCacheKey(email, linkID, queries)
	if err != nil {
		log.Printf("[stats] SetCachedStat hash error: %v", err)
		return
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("[stats] SetCachedStat marshal error: %v", err)
		return
	}

	r.Set(key, string(jsonData), r.ExpiredCache)
}

func InvalidateLinkCache(r *redis.Redis, linkID int64) {
	r.DelByPattern(fmt.Sprintf("*:link:%d:*", linkID))
}
