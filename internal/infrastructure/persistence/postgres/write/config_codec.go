package write

import (
	"encoding/json"
	"fmt"

	"github.com/petri-board-arena/internal/domain/arena"
)

type arenaConfigDTO struct {
	TickMillis         int     `json:"tickMillis"`
	Width              int     `json:"width"`
	Height             int     `json:"height"`
	DiffusionRate      float64 `json:"diffusionRate"`
	MutationRate       float64 `json:"mutationRate"`
	MaxOrganisms       int     `json:"maxOrganisms"`
	SnapshotEveryTicks int     `json:"snapshotEveryTicks"`
	Temperature        struct {
		Value float64 `json:"value"`
		Unit  string  `json:"unit"`
	} `json:"temperature"`
}

func ConfigFromJSON(b []byte) (arena.Config, error) {
	var dto arenaConfigDTO
	if err := json.Unmarshal(b, &dto); err != nil {
		return arena.Config{}, fmt.Errorf("unmarshal arena config: %w", err)
	}

	cfg := arena.Config{
		TickMillis:         dto.TickMillis,
		Width:              dto.Width,
		Height:             dto.Height,
		DiffusionRate:      dto.DiffusionRate,
		MutationRate:       dto.MutationRate,
		MaxOrganisms:       dto.MaxOrganisms,
		SnapshotEveryTicks: dto.SnapshotEveryTicks,
		Temperature: arena.Temperature{
			Value: dto.Temperature.Value,
			Unit:  arena.TemperatureUnit(dto.Temperature.Unit), // ajuste se for enum forte
		},
	}

	if err := cfg.Validate(); err != nil {
		return arena.Config{}, err
	}
	return cfg, nil
}
