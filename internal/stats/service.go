package stats

import (
	"adv/go-http/pkg/event"
	"log"
)

type StatServiceDep struct {
	EventBus        *event.EventBus
	StatsRepository *StatsRepository
}

type StatsService struct {
	EventBus        *event.EventBus
	StatsRepository *StatsRepository
}

func NewStatsService(deps *StatServiceDep) *StatsService {
	return &StatsService{
		EventBus:        deps.EventBus,
		StatsRepository: deps.StatsRepository,
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
