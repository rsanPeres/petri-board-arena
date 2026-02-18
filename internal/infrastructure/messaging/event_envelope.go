package messaging

import (
	"encoding/json"
	"time"
)

type EventEnvelope struct {
	EventID     string          `json:"eventId"`
	EventType   string          `json:"eventType"`
	AggregateID string          `json:"aggregateId"` // arenaId
	OccurredAt  time.Time       `json:"occurredAt"`
	Version     int             `json:"version"`
	Payload     json.RawMessage `json:"payload"`
}
