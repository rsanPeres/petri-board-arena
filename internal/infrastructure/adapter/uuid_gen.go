package adapter

import (
	"context"

	"github.com/google/uuid"
	"github.com/petri-board-arena/internal/domain/arena"
)

type UUIDGen struct{}

func (UUIDGen) NewArenaID(_ context.Context) (arena.ID, error) {
	return uuid.New(), nil
}
