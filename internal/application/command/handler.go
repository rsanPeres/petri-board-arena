package command

import (
	"context"
	"fmt"
	"strings"

	"github.com/petri-board-arena/internal/application/port"
	"github.com/petri-board-arena/internal/application/port/repository"
	"github.com/petri-board-arena/internal/domain/arena"
)

type IDGenerator interface {
	NewArenaID(ctx context.Context) (arena.ID, error)
}

type EventPublisher interface {
	Publish(ctx context.Context, events ...arena.Event) error
}

type Handler struct {
	uow    port.UnitOfWork
	repo   repository.ArenaWriteRepository
	ids    IDGenerator
	clock  port.Clock
	events EventPublisher
}

func NewHandler(
	uow port.UnitOfWork,
	repo repository.ArenaWriteRepository,
	ids IDGenerator,
	clock port.Clock,
	events EventPublisher,
) *Handler {
	return &Handler{
		uow:    uow,
		repo:   repo,
		ids:    ids,
		clock:  clock,
		events: events,
	}
}

func (h *Handler) Handle(ctx context.Context, cmd Command) (Result, error) {
	name := strings.TrimSpace(cmd.Name)
	if len(name) < 3 {
		return Result{}, fmt.Errorf("create_arena: name must have at least 3 chars")
	}

	if err := cmd.Config.Validate(); err != nil {
		return Result{}, fmt.Errorf("create_arena: invalid config: %w", err)
	}

	now := h.clock.Now()

	var out Result

	err := h.uow.WithinTransaction(ctx, func(txCtx context.Context) error {
		arenaID, err := h.ids.NewArenaID(txCtx)
		if err != nil {
			return fmt.Errorf("create_arena: generate id: %w", err)
		}

		a, err := arena.NewArena(arenaID, name, cmd.Config, now)
		if err != nil {
			return fmt.Errorf("create_arena: domain reject: %w", err)
		}

		if err := h.repo.Save(txCtx, a); err != nil {
			return fmt.Errorf("create_arena: persist: %w", err)
		}

		evs := a.PullEvents()
		if len(evs) > 0 {
			if err := h.events.Publish(txCtx, evs...); err != nil {
				return fmt.Errorf("create_arena: publish events: %w", err)
			}
		}

		out = Result{ArenaID: arenaID}
		return nil
	})

	if err != nil {
		return Result{}, err
	}
	return out, nil
}
