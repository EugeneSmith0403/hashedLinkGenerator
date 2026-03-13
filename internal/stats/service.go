package stats

import (
	"link-generator/pkg/event"
	"net"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/oschwald/geoip2-golang"
)

type StatServiceDep struct {
	EventBus        *event.EventBus
	StatsRepository *StatsRepository
	GeoIP           *geoip2.Reader
}

type StatsService struct {
	EventBus        *event.EventBus
	StatsRepository *StatsRepository
	geoIP           *geoip2.Reader
}

func NewStatsService(deps *StatServiceDep) *StatsService {
	return &StatsService{
		EventBus:        deps.EventBus,
		StatsRepository: deps.StatsRepository,
		geoIP:           deps.GeoIP,
	}
}

func (s *StatsService) BuildClientContext(r *http.Request) ClientContext {
	ip, port, _ := net.SplitHostPort(r.RemoteAddr)

	requestID := r.Header.Get("X-Request-ID")
	if requestID == "" {
		requestID = uuid.New().String()
	}

	country := s.lookupCountry(ip)

	ctx := ClientContext{
		IP:           ip,
		RemotePort:   port,
		RemoteAddr:   r.RemoteAddr,
		ForwardedFor: r.Header.Get("X-Forwarded-For"),
		RealIP:       r.Header.Get("X-Real-IP"),
		Country:      country,

		UserAgent:      r.UserAgent(),
		Accept:         r.Header.Get("Accept"),
		AcceptLanguage: r.Header.Get("Accept-Language"),
		AcceptEncoding: r.Header.Get("Accept-Encoding"),
		Origin:         r.Header.Get("Origin"),
		Referer:        r.Referer(),

		ForwardedProto: r.Header.Get("X-Forwarded-Proto"),
		ForwardedHost:  r.Header.Get("X-Forwarded-Host"),
		ForwardedPort:  r.Header.Get("X-Forwarded-Port"),

		RequestID: requestID,
		Timestamp: time.Now(),
	}

	if r.TLS != nil {
		ctx.Scheme = "https"
	} else {
		ctx.Scheme = "http"
	}

	return ctx
}

func (s *StatsService) lookupCountry(ip string) string {
	if s.geoIP == nil {
		return ""
	}
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return ""
	}
	record, err := s.geoIP.Country(parsed)
	if err != nil {
		return ""
	}
	return record.Country.IsoCode
}
