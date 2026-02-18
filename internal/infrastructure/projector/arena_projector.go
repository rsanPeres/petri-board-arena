package projector

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/petri-board-arena/internal/infrastructure/config"
	"github.com/petri-board-arena/internal/infrastructure/messaging"
	"github.com/redis/go-redis/v9"
)

type Projector struct {
	rdb *redis.Client
	cfg config.WorkerConfig
}

func NewProjector(rdb *redis.Client, cfg config.WorkerConfig) *Projector {
	return &Projector{rdb: rdb, cfg: cfg}
}

// Apply: roteia por EventType e escreve no Redis.
// IMPORTANT: mantém idempotência por eventId (SETNX).
func (p *Projector) Apply(ctx context.Context, ev messaging.EventEnvelope) error {
	if ev.EventID == "" || ev.EventType == "" || ev.AggregateID == "" {
		return errors.New("invalid event envelope: missing required fields")
	}

	// --- idempotência (at-least-once safe) ---
	idKey := "processed:event:" + ev.EventID
	ok, err := p.rdb.SetNX(ctx, idKey, "1", p.cfg.IdempotencyTTL).Result()
	if err != nil {
		return err
	}
	if !ok {
		// já processado
		return nil
	}

	switch ev.EventType {
	case "ArenaCreated":
		return p.onArenaCreated(ctx, ev)
	case "ArenaStarted":
		return p.onArenaStatus(ctx, ev, "RUNNING")
	case "ArenaPaused":
		return p.onArenaStatus(ctx, ev, "PAUSED")
	case "ArenaResumed":
		return p.onArenaStatus(ctx, ev, "RUNNING")
	case "ArenaStopped", "ArenaFinished":
		return p.onArenaStatus(ctx, ev, "FINISHED")
	default:
		// evento desconhecido: não falha o consumer (você pode logar/metricar)
		return nil
	}
}

// payload esperado (exemplo mínimo; ajuste para seu schema)
type arenaCreatedPayload struct {
	Name   string          `json:"name"`
	Config json.RawMessage `json:"config"`
}

func (p *Projector) onArenaCreated(ctx context.Context, ev messaging.EventEnvelope) error {
	var pl arenaCreatedPayload
	if err := json.Unmarshal(ev.Payload, &pl); err != nil {
		return err
	}
	if strings.TrimSpace(pl.Name) == "" {
		return errors.New("ArenaCreated payload missing name")
	}

	arenaKey := "arena:" + ev.AggregateID
	statusKey := "arenas:status:PENDING"
	createdZ := "arenas:created_at"

	createdAtScore := float64(ev.OccurredAt.Unix())

	pipe := p.rdb.TxPipeline()
	pipe.HSet(ctx, arenaKey, map[string]any{
		"id":         ev.AggregateID,
		"name":       pl.Name,
		"status":     "PENDING",
		"createdAt":  ev.OccurredAt.UTC().Format(time.RFC3339Nano),
		"updatedAt":  time.Now().UTC().Format(time.RFC3339Nano),
		"configJson": string(pl.Config), // guarda como string; alternativa: RedisJSON
	})
	pipe.SAdd(ctx, statusKey, ev.AggregateID)
	pipe.ZAdd(ctx, createdZ, redis.Z{Score: createdAtScore, Member: ev.AggregateID})

	_, err := pipe.Exec(ctx)
	return err
}

func (p *Projector) onArenaStatus(ctx context.Context, ev messaging.EventEnvelope, newStatus string) error {
	arenaKey := "arena:" + ev.AggregateID

	// lê status atual (para mover entre sets)
	oldStatus, _ := p.rdb.HGet(ctx, arenaKey, "status").Result()

	pipe := p.rdb.TxPipeline()
	pipe.HSet(ctx, arenaKey, map[string]any{
		"status":    newStatus,
		"updatedAt": time.Now().UTC().Format(time.RFC3339Nano),
	})

	if oldStatus != "" && oldStatus != newStatus {
		pipe.SRem(ctx, "arenas:status:"+oldStatus, ev.AggregateID)
	}
	pipe.SAdd(ctx, "arenas:status:"+newStatus, ev.AggregateID)

	_, err := pipe.Exec(ctx)
	return err
}
