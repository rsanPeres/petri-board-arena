package adapter

import (
	"context"

	"github.com/petri-board-arena/internal/domain/arena"
)

// substituir por Outbox/EventBus.
type NopArenaPublisher struct{}

func (NopArenaPublisher) Publish(_ context.Context, _ ...arena.Event) error { return nil }
