package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type txKey struct{}

func WithTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

func TxFromContext(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(pgx.Tx)
	return tx, ok
}

type TransactionManager struct {
	db *DB
}

func NewTransactionManager(db *DB) *TransactionManager {
	return &TransactionManager{db: db}
}

func (tm *TransactionManager) RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := tm.db.pool.Begin(ctx)
	if err != nil {
		return err
	}

	txCtx := WithTx(ctx, tx)

	if err := fn(txCtx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return errors.Join(err, rbErr)
		}
		return err
	}

	return tx.Commit(ctx)
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

func isNoRows(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}
