package stats

import (
	"adv/go-http/pkg/event"
	"adv/go-http/pkg/redis"
	"encoding/json"
	"log"
	"time"
)

type StatServiceDep struct {
	EventBus        *event.EventBus
	StatsRepository *StatsRepository
	RedisSrvice     *redis.Redis
}

type StatsService struct {
	EventBus        *event.EventBus
	StatsRepository *StatsRepository
	redisSrvice     *redis.Redis
}

func NewStatsService(deps *StatServiceDep) *StatsService {
	return &StatsService{
		EventBus:        deps.EventBus,
		StatsRepository: deps.StatsRepository,
		redisSrvice:     deps.RedisSrvice,
	}
}

func (s *StatsService) AddClick() {
	for msg := range s.EventBus.Subscribe() {
		if msg.Type == event.LinkVisitedEVent {
			data, ok := msg.Data.(int)

			if !ok {
				log.Fatalln("Bad LinkVisitedEVent data:", msg.Data)
				continue
			}
			s.StatsRepository.UpdateLinkClicks(data)
		}

	}
}

func GetCachedStat[T any](s *StatsService, queries map[string]time.Time) (T, error) {
	var zero T

	hashQuery, err := redis.HashFilters(queries)
	if err != nil {
		return zero, err
	}

	statsData := s.redisSrvice.Get(hashQuery)
	if statsData == "" {
		return zero, nil
	}

	var result T
	if err := json.Unmarshal([]byte(statsData), &result); err != nil {
		return zero, err
	}

	return result, nil
}

func SetCachedStat[T any](s *StatsService, data T, queries map[string]time.Time) {
	key, err := redis.HashFilters(queries)
	if err != nil {
		log.Printf("[stats] SetCachedStat hash error: %v", err)
		return
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("[stats] SetCachedStat marshal error: %v", err)
		return
	}

	s.redisSrvice.Set(key, string(jsonData), s.redisSrvice.ExpiredCache)
}
