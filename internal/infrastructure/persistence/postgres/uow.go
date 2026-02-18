package postgres

import (
	"context"
	"database/sql"
)

type UnitOfWork struct {
	db *sql.DB
}

func NewUnitOfWork(db *sql.DB) *UnitOfWork { return &UnitOfWork{db: db} }

func (u *UnitOfWork) WithinTransaction(ctx context.Context, fn func(txCtx context.Context) error) error {
	tx, err := u.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}

	txCtx := WithTx(ctx, tx)

	if err := fn(txCtx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}
