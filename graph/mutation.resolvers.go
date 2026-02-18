package graph

import (
	"context"
	"fmt"

	"github.com/petri-board-arena/graph/model"
	createarena "github.com/petri-board-arena/internal/application/command"
	"github.com/petri-board-arena/internal/domain/arena"
)

type mutationResolver struct{ r *Resolver }

func (m *mutationResolver) CreateArena(ctx context.Context, input model.CreateArenaInput) (*model.CreateArenaPayload, error) {
	cfg, err := mapArenaConfigInputToDomain(input.Config)
	if err != nil {
		return nil, err
	}

	cmd := createarena.Command{
		Name:   input.Name,
		Config: cfg,
	}

	res, err := m.r.CreateArenaHandler.Handle(ctx, cmd)
	if err != nil {
		return nil, err
	}

	return &model.CreateArenaPayload{
		Arena: &model.Arena{
			ID: res.ArenaID,
		},
	}, nil
}

func mapArenaConfigInputToDomain(in *model.ArenaConfigInput) (arena.Config, error) {
	if in == nil {
		return arena.Config{}, fmt.Errorf("config is required")
	}
	if in.Temperature == nil {
		return arena.Config{}, fmt.Errorf("temperature is required")
	}

	cfg := arena.Config{
		TickMillis:         int(in.TickMillis),
		Width:              int(in.Width),
		Height:             int(in.Height),
		DiffusionRate:      in.DiffusionRate,
		MutationRate:       in.MutationRate,
		MaxOrganisms:       int(in.MaxOrganisms),
		SnapshotEveryTicks: int(in.SnapshotEveryTicks),
		Temperature: arena.Temperature{
			Value: in.Temperature.Value,
			Unit:  arena.TemperatureUnit(in.Temperature.Unit),
		},
	}

	if err := cfg.Validate(); err != nil {
		return arena.Config{}, err
	}

	return cfg, nil
}
