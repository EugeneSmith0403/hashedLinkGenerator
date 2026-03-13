package stats

import (
	"encoding/json"
	"fmt"
	"link-generator/pkg/redis"
	"log"
	"time"
)

func statCacheKey(linkID uint, queries map[string]time.Time) (string, error) {
	hashQuery, err := redis.HashFilters(queries)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("link:%d:report:%s", linkID, hashQuery), nil
}

func GetCachedStat[T any](r *redis.Redis, queries map[string]time.Time, linkID uint) (T, error) {
	var zero T

	key, err := statCacheKey(linkID, queries)
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

func SetCachedStat[T any](r *redis.Redis, data T, queries map[string]time.Time, linkID uint) {
	key, err := statCacheKey(linkID, queries)
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

func InvalidateLinkCache(r *redis.Redis, linkID uint, hashQuery string) (int64, error) {
	key := fmt.Sprintf("link:%d:report:%s", linkID, hashQuery)

	return r.Del(key), nil
}
