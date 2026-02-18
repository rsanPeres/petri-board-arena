package arena

import (
	"fmt"
	"time"
)

type Player struct {
	ID          PlayerID
	DisplayName string
	Role        PlayerRole
	JoinedAt    time.Time
}

type PlayerAction struct {
	ID          ActionID
	Type        ActionType
	PlayerID    PlayerID
	SubmittedAt time.Time
	ApplyAtTick int64
	Payload     ActionPayload
}

type ActionPayload interface {
	isPayload()
	Validate(cfg Config) error
}

type AddNutrientsPayload struct {
	Area   Area
	Amount int
}

func (AddNutrientsPayload) isPayload() {}
func (p AddNutrientsPayload) Validate(cfg Config) error {
	if p.Amount <= 0 {
		return fmt.Errorf("%w: amount must be > 0", ErrInvalidAction)
	}
	return p.Area.Validate(cfg.Width, cfg.Height)
}

type DropAntibioticPayload struct {
	Area          Area
	Kind          AntibioticKind
	Concentration float64
}

func (DropAntibioticPayload) isPayload() {}
func (p DropAntibioticPayload) Validate(cfg Config) error {
	if p.Concentration <= 0 {
		return fmt.Errorf("%w: concentration must be > 0", ErrInvalidAction)
	}
	if p.Kind != AntibioticA && p.Kind != AntibioticB && p.Kind != AntibioticC {
		return fmt.Errorf("%w: invalid antibiotic kind", ErrInvalidAction)
	}
	return p.Area.Validate(cfg.Width, cfg.Height)
}

type SetTemperaturePayload struct {
	Temperature Temperature
}

func (SetTemperaturePayload) isPayload() {}
func (p SetTemperaturePayload) Validate(cfg Config) error {
	return p.Temperature.Validate()
}

type SpawnOrganismPayload struct {
	Kind             OrganismKind
	Position         Point
	GenomeTemplateID *string // opcional no domínio
}

func (SpawnOrganismPayload) isPayload() {}
func (p SpawnOrganismPayload) Validate(cfg Config) error {
	if p.Kind != KindBacteria && p.Kind != KindFungi && p.Kind != KindPhage {
		return fmt.Errorf("%w: invalid organism kind", ErrInvalidAction)
	}
	return p.Position.Validate(cfg.Width, cfg.Height)
}
