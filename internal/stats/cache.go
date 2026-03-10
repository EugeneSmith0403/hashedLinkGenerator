package stats

import (
	"link-generator/pkg/redis"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

func GetCachedStat[T any](r *redis.Redis, queries map[string]time.Time, email string) (T, error) {
	var zero T

	hashQuery, err := redis.HashFilters(queries)
	if err != nil {
		return zero, err
	}

	key := fmt.Sprintf("email:%sreport:%s", email, hashQuery)

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

func SetCachedStat[T any](r *redis.Redis, data T, queries map[string]time.Time, email string) {
	key, err := redis.HashFilters(queries)
	if err != nil {
		log.Printf("[stats] SetCachedStat hash error: %v", err)
		return
	}

	key = fmt.Sprintf("email:%sreport:%s", email, key)

	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("[stats] SetCachedStat marshal error: %v", err)
		return
	}

	r.Set(key, string(jsonData), r.ExpiredCache)
}
