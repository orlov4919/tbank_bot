package transactor

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Transactor struct {
	pool *pgxpool.Pool
}

func New(pgxPool *pgxpool.Pool) *Transactor {
	return &Transactor{
		pool: pgxPool,
	}
}

type Querier interface {
	Exec(ctx context.Context, sql string, arguments ...any) (commandTag pgconn.CommandTag, err error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func GetQuerier(ctx context.Context, defaultQuerier Querier) Querier {
	if querier := ctx.Value(txKey{}); querier != nil {
		return querier.(pgx.Tx)
	}

	return defaultQuerier
}

type txKey struct{}

func (t *Transactor) InjectTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

func (t *Transactor) WithTransaction(ctx context.Context, txFunc func(ctx context.Context) error) (err error) {
	tx, err := t.pool.Begin(ctx)

	if err != nil {
		return fmt.Errorf("ошибка при старте транзакции: %w", err)
	}

	defer func() {
		if err != nil {
			err = errors.Join(err, tx.Rollback(ctx))
		}
	}()

	if err = txFunc(t.InjectTx(ctx, tx)); err != nil {
		return fmt.Errorf("ошибка при выполнении транзакции: %w", err)
	}

	return tx.Commit(ctx)
}
