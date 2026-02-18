package repository

import (
	"context"

	"github.com/petri-board-arena/internal/application/query/dto"
)

type ArenaReadRepository interface {
	GetArena(ctx context.Context, id string) (*dto.ArenaView, error)
	ListArenas(ctx context.Context, status *string, limit, offset int) ([]dto.ArenaView, int, error)
}
