package repository

import (
	"context"

	"github.com/petri-board-arena/internal/domain/arena"
)

type ArenaWriteRepository interface {
	GetByID(ctx context.Context, id arena.ID) (*arena.Arena, error)
	Save(ctx context.Context, a *arena.Arena) error
}
