package arena

import (
	"time"
)

type Event interface {
	EventName() string
	OccurredAt() time.Time
	ArenaID() ID
}

type baseEvent struct {
	at      time.Time
	arenaID ID
}

func (b baseEvent) OccurredAt() time.Time { return b.at }
func (b baseEvent) ArenaID() ID           { return b.arenaID }

type ArenaCreated struct {
	baseEvent
	Name   string
	Config Config
}

func (e ArenaCreated) EventName() string { return "ArenaCreated" }

type ArenaStarted struct{ baseEvent }

func (e ArenaStarted) EventName() string { return "ArenaStarted" }

type ArenaPaused struct{ baseEvent }

func (e ArenaPaused) EventName() string { return "ArenaPaused" }

type ArenaResumed struct{ baseEvent }

func (e ArenaResumed) EventName() string { return "ArenaResumed" }

type ArenaStopped struct{ baseEvent }

func (e ArenaStopped) EventName() string { return "ArenaStopped" }

type PlayerJoined struct {
	baseEvent
	PlayerID    PlayerID
	DisplayName string
	Role        PlayerRole
}

func (e PlayerJoined) EventName() string { return "PlayerJoined" }

type PlayerLeft struct {
	baseEvent
	PlayerID PlayerID
}

func (e PlayerLeft) EventName() string { return "PlayerLeft" }

type ArenaConfigUpdated struct {
	baseEvent
	Config Config
}

func (e ArenaConfigUpdated) EventName() string { return "ArenaConfigUpdated" }

type ActionSubmitted struct {
	baseEvent
	Action PlayerAction
}

func (e ActionSubmitted) EventName() string { return "ActionSubmitted" }

type TickAdvanced struct {
	baseEvent
	Tick int64
}

func (e TickAdvanced) EventName() string { return "TickAdvanced" }
