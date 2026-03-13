package models

import "time"

type StatsEventType string

const (
	StatsLinkVisited StatsEventType = "link.visited"
)

type StatsData struct {
	LinkHash string `json:"linkHash"`
	UserID   uint   `json:"userId"`
	IP       string `json:"ip"`
}

type StatsMessage struct {
	EventType StatsEventType         `json:"eventType"`
	Data      *LinkTransitionWitHash `json:"data"`
}

type LinkTransition struct {
	LinkID    int64     `ch:"link_id"`
	ClickedAt time.Time `ch:"clicked_at"`

	// Network
	IP           string `ch:"ip"`
	ForwardedFor string `ch:"forwarded_for"`
	RealIP       string `ch:"real_ip"`
	RemoteAddr   string `ch:"remote_addr"`
	RemotePort   string `ch:"remote_port"`
	Country      string `ch:"country"`

	// Headers
	UserAgent      string `ch:"user_agent"`
	Accept         string `ch:"accept"`
	AcceptLanguage string `ch:"accept_language"`
	AcceptEncoding string `ch:"accept_encoding"`
	Origin         string `ch:"origin"`
	Referer        string `ch:"referer"`

	// Device
	DeviceType string `ch:"device_type"`
	OS         string `ch:"os"`
	Browser    string `ch:"browser"`

	// Security
	Fingerprint    string `ch:"fingerprint"`
	RequestID      string `ch:"request_id"`
	ForwardedProto string `ch:"forwarded_proto"`
	ForwardedHost  string `ch:"forwarded_host"`
	ForwardedPort  string `ch:"forwarded_port"`
	Scheme         string `ch:"scheme"`
}

type LinkTransitionWitHash struct {
	*LinkTransition
	FilterHash string
}
