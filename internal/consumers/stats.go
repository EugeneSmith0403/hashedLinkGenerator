package consumers

import (
	"encoding/json"
	"fmt"
	"log"
	"slices"

	"link-generator/internal/models"
	"link-generator/internal/stats"
	"link-generator/pkg/redis"
)

type StatsConsumer struct {
	statsRepo *stats.StatsRepository
	redis     *redis.Redis
}

func NewStatsConsumer(statsRepo *stats.StatsRepository, redis *redis.Redis) *StatsConsumer {
	return &StatsConsumer{
		statsRepo: statsRepo,
		redis:     redis,
	}
}

func (sc *StatsConsumer) Handle(body []byte) error {
	var msg models.StatsMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	switch msg.EventType {
	case models.StatsLinkVisited:
		return sc.handleLinkVisited(msg.Data)
	default:
		log.Printf("[consumer] no handler for %s, skipping LinkID=%d", msg.EventType, msg.Data.LinkID)
		return nil
	}
}

func (sc *StatsConsumer) handleLinkVisited(data *models.LinkTransitionWitHash) error {
	list := []models.LinkTransition{*data.LinkTransition}
	list = slices.DeleteFunc(list, func(t models.LinkTransition) bool {
		return t.LinkID == 0
	})
	if err := sc.statsRepo.Insert(list); err != nil {
		return err
	}

	stats.InvalidateLinkCache(sc.redis, uint(data.LinkID), data.FilterHash)

	return nil
}
