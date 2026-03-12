package models

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
	EventType StatsEventType `json:"eventType"`
	Data      StatsData      `json:"data"`
}
