package transactor

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Connection interface {
	Exec(ctx context.Context, sql string, args ...any) (commandTag pgconn.CommandTag, err error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
}

type database interface {
	Connection
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
}

type txKey struct{}

type Transactor struct {
	database  database
	txOptions pgx.TxOptions
}

func (tr *Transactor) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := tr.database.BeginTx(ctx, tr.txOptions)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}

	defer func() {
		if err = tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			slog.ErrorContext(ctx, "failed to rollback tx", slog.String("error", err.Error()))
		}
	}()

	if err = fn(context.WithValue(ctx, txKey{}, tx)); err != nil {
		return fmt.Errorf("failed to apply tx: %w", err)
	}
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit tx: %w", err)
	}

	return nil
}

func (tr *Transactor) Conn(ctx context.Context) Connection { //nolint:ireturn
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}
	return tr.database
}

func NewTransactor(database database, txOptions pgx.TxOptions) *Transactor {
	return &Transactor{
		database:  database,
		txOptions: txOptions,
	}
}
