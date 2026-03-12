package stats

import (
	"link-generator/pkg/event"
	"log"
	"net"
	"net/http"
	"time"
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

func (s *StatsService) BuildClientContext(r *http.Request) ClientContext {

	ip, port, _ := net.SplitHostPort(r.RemoteAddr)

	ctx := ClientContext{
		IP:           ip,
		RemotePort:   port,
		RemoteAddr:   r.RemoteAddr,
		ForwardedFor: r.Header.Get("X-Forwarded-For"),
		RealIP:       r.Header.Get("X-Real-IP"),

		UserAgent:      r.UserAgent(),
		Accept:         r.Header.Get("Accept"),
		AcceptLanguage: r.Header.Get("Accept-Language"),
		AcceptEncoding: r.Header.Get("Accept-Encoding"),
		Origin:         r.Header.Get("Origin"),
		Referer:        r.Referer(),

		ForwardedProto: r.Header.Get("X-Forwarded-Proto"),
		ForwardedHost:  r.Header.Get("X-Forwarded-Host"),
		ForwardedPort:  r.Header.Get("X-Forwarded-Port"),

		Timestamp: time.Now(),
	}

	if r.TLS != nil {
		ctx.Scheme = "https"
	} else {
		ctx.Scheme = "http"
	}

	return ctx
}
