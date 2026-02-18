package read

import (
	"context"
	"database/sql"

	"github.com/petri-board-arena/internal/application/query/dto"
)

type ArenaReadRepo struct {
	db *sql.DB
}

func NewArenaReadRepo(db *sql.DB) *ArenaReadRepo { return &ArenaReadRepo{db: db} }

func (r *ArenaReadRepo) GetArena(ctx context.Context, id string) (*dto.ArenaView, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, name, status, tick, created_at, started_at
		FROM arena_read
		WHERE id = $1
	`, id)

	var v dto.ArenaView
	if err := row.Scan(&v.ID, &v.Name, &v.Status, &v.Tick, &v.CreatedAt, &v.StartedAt); err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *ArenaReadRepo) ListArenas(ctx context.Context, status *string, limit, offset int) ([]dto.ArenaView, int, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, status, tick, created_at, started_at
		FROM arena_read
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]dto.ArenaView, 0, limit)
	for rows.Next() {
		var v dto.ArenaView
		if err := rows.Scan(&v.ID, &v.Name, &v.Status, &v.Tick, &v.CreatedAt, &v.StartedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, v)
	}

	// total: em produção, faça COUNT separado
	total := len(out)
	return out, total, rows.Err()
}
