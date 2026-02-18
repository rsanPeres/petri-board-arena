package graph

import (
	"context"
	"time"

	"github.com/petri-board-arena/graph/model"
)

type queryResolver struct{ r *Resolver }

func (q *queryResolver) Health(ctx context.Context) (*model.Health, error) {
	now := time.Now().UTC()
	return &model.Health{
		Ok:      true,
		Version: "dev",
		Now:     now,
	}, nil
}
