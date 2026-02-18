package port

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type OutboxStatus string

const (
	OutboxPending    OutboxStatus = "PENDING"
	OutboxProcessing OutboxStatus = "PROCESSING"
	OutboxPublished  OutboxStatus = "PUBLISHED"
	OutboxFailed     OutboxStatus = "FAILED"
)

// OutboxEvent maps to table: outbox_event
type OutboxEvent struct {
	ID uuid.UUID `json:"id"`

	// Identity / routing
	AggregateType string `json:"aggregateType"`
	AggregateID   string `json:"aggregateId"`
	EventType     string `json:"eventType"`
	Topic         string `json:"topic"`

	// Payload
	Payload json.RawMessage `json:"payload"`
	Headers json.RawMessage `json:"headers"`

	// Traceability / idempotency
	CorrelationID  *string `json:"correlationId,omitempty"`
	CausationID    *string `json:"causationId,omitempty"`
	IdempotencyKey *string `json:"idempotencyKey,omitempty"`

	// Publish state
	Status      OutboxStatus `json:"status"`
	Attempts    int          `json:"attempts"`
	MaxAttempts int          `json:"maxAttempts"`

	NextAttemptAt time.Time  `json:"nextAttemptAt"`
	PublishedAt   *time.Time `json:"publishedAt,omitempty"`

	// Locking (multi outbox workers)
	LockedBy      *string    `json:"lockedBy,omitempty"`
	LockedAt      *time.Time `json:"lockedAt,omitempty"`
	LockExpiresAt *time.Time `json:"lockExpiresAt,omitempty"`

	// Audit
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type OutboxEnqueueParams struct {
	ID uuid.UUID

	AggregateType string
	AggregateID   string
	EventType     string
	Topic         string

	Payload json.RawMessage
	Headers json.RawMessage

	CorrelationID  *string
	CausationID    *string
	IdempotencyKey *string

	MaxAttempts   int
	NextAttemptAt time.Time
}

// OutboxLockParams controls polling/locking behavior.
type OutboxLockParams struct {
	WorkerID  string
	BatchSize int
	LockTTL   time.Duration
}

type OutboxMarkFailedParams struct {
	EventID      uuid.UUID
	BaseBackoff  time.Duration
	LastErrorMsg string
}

type OutboxDeadLetter struct {
	ID uuid.UUID `json:"id"`

	OutboxEventID *uuid.UUID `json:"outboxEventId,omitempty"`

	AggregateType string `json:"aggregateType"`
	AggregateID   string `json:"aggregateId"`
	EventType     string `json:"eventType"`
	Topic         string `json:"topic"`

	Payload json.RawMessage `json:"payload"`
	Headers json.RawMessage `json:"headers"`

	CorrelationID  *string `json:"correlationId,omitempty"`
	CausationID    *string `json:"causationId,omitempty"`
	IdempotencyKey *string `json:"idempotencyKey,omitempty"`

	Attempts  int    `json:"attempts"`
	LastError string `json:"lastError"`

	CreatedAt time.Time `json:"createdAt"`
}

type OutboxDeadLetterParams struct {
	ID uuid.UUID

	OutboxEventID *uuid.UUID

	AggregateType string
	AggregateID   string
	EventType     string
	Topic         string

	Payload json.RawMessage
	Headers json.RawMessage

	CorrelationID  *string
	CausationID    *string
	IdempotencyKey *string

	Attempts  int
	LastError string
}
