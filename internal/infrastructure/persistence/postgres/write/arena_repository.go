package write

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	uuid "github.com/google/uuid"

	"github.com/petri-board-arena/internal/domain/arena"
	"github.com/petri-board-arena/internal/infrastructure/persistence/postgres"
)

var ErrArenaNotFound = errors.New("arena not found")

type ArenaRepo struct {
	db *sql.DB
}

func NewArenaRepo(db *sql.DB) *ArenaRepo { return &ArenaRepo{db: db} }

func (r *ArenaRepo) GetByID(ctx context.Context, id arena.ID) (*arena.Arena, error) {
	if tx, ok := postgres.TxFrom(ctx); ok {
		return r.getByID(ctx, tx, id, true)
	}
	return r.getByID(ctx, r.db, id, false)
}

type queryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

func (r *ArenaRepo) getByID(ctx context.Context, q queryer, id arena.ID, forUpdate bool) (*arena.Arena, error) {
	lock := ""
	if forUpdate {
		lock = " FOR UPDATE"
	}

	row := q.QueryRowContext(ctx, `
		SELECT
			id,
			name,
			status,
			created_at,
			started_at,
			finished_at,
			tick,
			config_json
		FROM arenas
		WHERE id = $1`+lock, id,
	)

	var (
		rawID      uuid.UUID
		name       string
		statusStr  string
		createdAt  time.Time
		startedAt  sql.NullTime
		finishedAt sql.NullTime
		tick       int64
		configJSON []byte
	)

	if err := row.Scan(
		&rawID,
		&name,
		&statusStr,
		&createdAt,
		&startedAt,
		&finishedAt,
		&tick,
		&configJSON,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrArenaNotFound
		}
		return nil, err
	}

	cfg, err := ConfigFromJSON(configJSON)
	if err != nil {
		return nil, fmt.Errorf("decode config_json: %w", err)
	}

	st, err := arena.ParseStatus(statusStr)
	if err != nil {
		return nil, fmt.Errorf("parse status: %w", err)
	}

	var (
		sAt *time.Time
		fAt *time.Time
	)
	if startedAt.Valid {
		t := startedAt.Time.UTC()
		sAt = &t
	}
	if finishedAt.Valid {
		t := finishedAt.Time.UTC()
		fAt = &t
	}

	a, err := arena.Rehydrate(arena.RehydrateState{
		ID:         rawID,
		Name:       name,
		Status:     st,
		CreatedAt:  createdAt.UTC(),
		StartedAt:  sAt,
		FinishedAt: fAt,
		Tick:       tick,
		Config:     cfg,
		Players:    nil,
		Scheduled:  nil,
	})
	if err != nil {
		return nil, err
	}

	return a, nil
}

func (r *ArenaRepo) Save(ctx context.Context, a *arena.Arena) error {
	if tx, ok := postgres.TxFrom(ctx); ok {
		return r.save(ctx, tx, a)
	}
	return r.save(ctx, r.db, a)
}

func (r *ArenaRepo) save(ctx context.Context, q queryer, a *arena.Arena) error {
	_, err := q.ExecContext(ctx, `
		INSERT INTO arenas (id, name, status, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET
		  name = EXCLUDED.name,
		  status = EXCLUDED.status
	`, a.ID(), a.Name(), a.Status(), a.CreatedAt())
	return err
}
