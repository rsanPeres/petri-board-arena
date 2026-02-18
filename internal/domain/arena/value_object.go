package arena

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// ----------------------------
// Types / Value Objects
// ----------------------------

type PlayerID uuid.UUID
type ActionID uuid.UUID

type Status string

const (
	StatusPending  Status = "PENDING"
	StatusRunning  Status = "RUNNING"
	StatusPaused   Status = "PAUSED"
	StatusFinished Status = "FINISHED"
)

type PlayerRole string

const (
	RoleAdmin  PlayerRole = "ADMIN"
	RolePlayer PlayerRole = "PLAYER"
)

type ActionType string

const (
	ActionAddNutrients   ActionType = "ADD_NUTRIENTS"
	ActionDropAntibiotic ActionType = "DROP_ANTIBIOTIC"
	ActionSetTemperature ActionType = "SET_TEMPERATURE"
	ActionSpawnOrganism  ActionType = "SPAWN_ORGANISM"
)

type OrganismKind string

const (
	KindBacteria OrganismKind = "BACTERIA"
	KindFungi    OrganismKind = "FUNGI"
	KindPhage    OrganismKind = "PHAGE"
)

type AntibioticKind string

const (
	AntibioticA AntibioticKind = "A"
	AntibioticB AntibioticKind = "B"
	AntibioticC AntibioticKind = "C"
)

type TemperatureUnit string

const (
	TempC TemperatureUnit = "C"
)

type Temperature struct {
	Value float64
	Unit  TemperatureUnit
}

func (t Temperature) Validate() error {
	if t.Unit != TempC {
		return fmt.Errorf("unsupported temperature unit: %s", t.Unit)
	}
	if t.Value < -50 || t.Value > 150 {
		return fmt.Errorf("temperature out of range: %v", t.Value)
	}
	return nil
}

type Point struct {
	X int
	Y int
}

func (p Point) Validate(width, height int) error {
	if p.X < 0 || p.Y < 0 || p.X >= width || p.Y >= height {
		return fmt.Errorf("point out of bounds (%d,%d) for grid %dx%d", p.X, p.Y, width, height)
	}
	return nil
}

type Area struct {
	X      int
	Y      int
	Width  int
	Height int
}

func (a Area) Validate(gridW, gridH int) error {
	if a.Width <= 0 || a.Height <= 0 {
		return errors.New("area width/height must be > 0")
	}
	if a.X < 0 || a.Y < 0 || a.X+a.Width > gridW || a.Y+a.Height > gridH {
		return fmt.Errorf("area out of bounds (%d,%d %dx%d) for grid %dx%d", a.X, a.Y, a.Width, a.Height, gridW, gridH)
	}
	return nil
}

type Config struct {
	TickMillis         int
	Width              int
	Height             int
	DiffusionRate      float64
	MutationRate       float64
	MaxOrganisms       int
	SnapshotEveryTicks int
	Temperature        Temperature
}

func (c Config) Validate() error {
	if c.TickMillis <= 0 {
		return errors.New("tickMillis must be > 0")
	}
	if c.Width <= 0 || c.Height <= 0 {
		return errors.New("width/height must be > 0")
	}
	if c.DiffusionRate < 0 || c.DiffusionRate > 1 {
		return errors.New("diffusionRate must be in [0,1]")
	}
	if c.MutationRate < 0 || c.MutationRate > 1 {
		return errors.New("mutationRate must be in [0,1]")
	}
	if c.MaxOrganisms <= 0 {
		return errors.New("maxOrganisms must be > 0")
	}
	if c.SnapshotEveryTicks <= 0 {
		return errors.New("snapshotEveryTicks must be > 0")
	}
	if err := c.Temperature.Validate(); err != nil {
		return err
	}
	return nil
}

func ParseStatus(s string) (Status, error) {
	switch Status(s) {
	case StatusPending, StatusRunning, StatusPaused, StatusFinished:
		return Status(s), nil
	default:
		return "", fmt.Errorf("invalid arena status: %q", s)
	}
}
