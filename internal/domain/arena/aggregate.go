package arena

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ID = uuid.UUID

type Arena struct {
	id         ID
	name       string
	status     Status
	createdAt  time.Time
	startedAt  *time.Time
	finishedAt *time.Time

	tick   int64
	config Config

	players          map[PlayerID]Player
	scheduledActions map[int64][]PlayerAction

	events []Event
}

// ----- Getters

func (a *Arena) ID() ID                 { return a.id }
func (a *Arena) Name() string           { return a.name }
func (a *Arena) Status() Status         { return a.status }
func (a *Arena) Tick() int64            { return a.tick }
func (a *Arena) Config() Config         { return a.config }
func (a *Arena) CreatedAt() time.Time   { return a.createdAt }
func (a *Arena) StartedAt() *time.Time  { return a.startedAt }
func (a *Arena) FinishedAt() *time.Time { return a.finishedAt }

func (a *Arena) Players() []Player {
	out := make([]Player, 0, len(a.players))
	for _, p := range a.players {
		out = append(out, p)
	}
	return out
}

func (a *Arena) PullEvents() []Event {
	ev := a.events
	a.events = nil
	return ev
}

func (a *Arena) record(e Event) { a.events = append(a.events, e) }

// ----------------------------
// Factories
// ----------------------------

func NewArena(id ID, name string, cfg Config, now time.Time) (*Arena, error) {
	if id == uuid.Nil {
		return nil, errors.New("arena id is required")
	}
	if len(name) < 3 {
		return nil, ErrInvalidName
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	ar := &Arena{
		id:               id,
		name:             name,
		status:           StatusPending,
		createdAt:        now.UTC(),
		tick:             0,
		config:           cfg,
		players:          make(map[PlayerID]Player),
		scheduledActions: make(map[int64][]PlayerAction),
	}

	ar.record(ArenaCreated{
		baseEvent: baseEvent{at: now.UTC(), arenaID: id},
		Name:      name,
		Config:    cfg,
	})

	return ar, nil
}

// Rehydrate for repository use (no events)
type RehydrateState struct {
	ID         ID
	Name       string
	Status     Status
	CreatedAt  time.Time
	StartedAt  *time.Time
	FinishedAt *time.Time
	Tick       int64
	Config     Config
	Players    []Player
	Scheduled  map[int64][]PlayerAction
}

func Rehydrate(s RehydrateState) (*Arena, error) {
	if s.ID == uuid.Nil || s.Name == "" {
		return nil, errors.New("rehydrate: missing required fields")
	}
	if err := s.Config.Validate(); err != nil {
		return nil, err
	}
	if s.Status != StatusPending && s.Status != StatusRunning && s.Status != StatusPaused && s.Status != StatusFinished {
		return nil, fmt.Errorf("rehydrate: invalid status: %s", s.Status)
	}

	ar := &Arena{
		id:               s.ID,
		name:             s.Name,
		status:           s.Status,
		createdAt:        s.CreatedAt,
		startedAt:        s.StartedAt,
		finishedAt:       s.FinishedAt,
		tick:             s.Tick,
		config:           s.Config,
		players:          make(map[PlayerID]Player),
		scheduledActions: make(map[int64][]PlayerAction),
	}

	for _, p := range s.Players {
		ar.players[p.ID] = p
	}
	if s.Scheduled != nil {
		ar.scheduledActions = s.Scheduled
	}

	return ar, nil
}

// ----------------------------
// Domain behavior
// ----------------------------

func (a *Arena) Start(now time.Time, by PlayerID) error {
	if a.status == StatusFinished {
		return ErrArenaFinished
	}
	if a.status != StatusPending {
		return ErrArenaNotPending
	}
	if by != PlayerID(uuid.Nil) && !a.isAdmin(by) {
		return ErrPermissionDenied
	}

	n := now.UTC()
	a.status = StatusRunning
	a.startedAt = &n
	a.record(ArenaStarted{baseEvent: baseEvent{at: n, arenaID: a.id}})
	return nil
}

func (a *Arena) Pause(now time.Time, by PlayerID) error {
	if a.status == StatusFinished {
		return ErrArenaFinished
	}
	if a.status != StatusRunning {
		return ErrArenaNotRunning
	}
	if by != PlayerID(uuid.Nil) && !a.isAdmin(by) {
		return ErrPermissionDenied
	}

	n := now.UTC()
	a.status = StatusPaused
	a.record(ArenaPaused{baseEvent: baseEvent{at: n, arenaID: a.id}})
	return nil
}

func (a *Arena) Resume(now time.Time, by PlayerID) error {
	if a.status == StatusFinished {
		return ErrArenaFinished
	}
	if a.status != StatusPaused {
		return ErrArenaNotPaused
	}
	if by != PlayerID(uuid.Nil) && !a.isAdmin(by) {
		return ErrPermissionDenied
	}

	n := now.UTC()
	a.status = StatusRunning
	a.record(ArenaResumed{baseEvent: baseEvent{at: n, arenaID: a.id}})
	return nil
}

func (a *Arena) Stop(now time.Time, by PlayerID) error {
	if a.status == StatusFinished {
		return nil
	}
	if by != PlayerID(uuid.Nil) && !a.isAdmin(by) {
		return ErrPermissionDenied
	}

	n := now.UTC()
	a.status = StatusFinished
	a.finishedAt = &n
	a.record(ArenaStopped{baseEvent: baseEvent{at: n, arenaID: a.id}})
	return nil
}

// ----------------------------
// Helpers
// ----------------------------

func (a *Arena) isAdmin(pid PlayerID) bool {
	p, ok := a.players[pid]
	return ok && p.Role == RoleAdmin
}
