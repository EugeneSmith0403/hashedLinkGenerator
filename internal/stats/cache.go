package stats

import (
	"encoding/json"
	"fmt"
	"link-generator/pkg/redis"
	"log"
	"time"
)

func statCacheKey(linkHash string, queries map[string]time.Time) (string, error) {
	hashQuery, err := redis.HashFilters(queries)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("link:%s:report:%s", linkHash, hashQuery), nil
}

func linkFilterSetKey(linkHash string) string {
	return fmt.Sprintf("link:%s:filters", linkHash)
}

func GetCachedStat[T any](r *redis.Redis, queries map[string]time.Time, linkHash string) (T, error) {
	var zero T

	key, err := statCacheKey(linkHash, queries)
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

func SetCachedStat[T any](r *redis.Redis, data T, queries map[string]time.Time, linkHash string) {
	key, err := statCacheKey(linkHash, queries)
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
	r.SAdd(linkFilterSetKey(linkHash), key)
}

func InvalidateLinkCache(r *redis.Redis, linkID uint, linkHash string) (int64, error) {
	setKey := linkFilterSetKey(linkHash)
	keys := r.SMembers(setKey)

	var deleted int64
	if len(keys) > 0 {
		deleted = r.Del(keys...)
	}
	r.Del(setKey)

	return deleted, nil
}
