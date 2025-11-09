package transactor

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
)

type Connection interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type database interface {
	Connection
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

type txKey struct{}

type Transactor struct {
	database  database
	txOptions *sql.TxOptions
}

func (tr *Transactor) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := tr.database.BeginTx(ctx, tr.txOptions)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %s", err)
	}

	defer func() {
		if err = tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			slog.ErrorContext(ctx, "failed to rollback tx", slog.String("error", err.Error()))
		}
	}()

	if err = fn(context.WithValue(ctx, txKey{}, tx)); err != nil {
		return fmt.Errorf("failed to apply tx: %s", err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit tx: %s", err)
	}

	return nil
}

func (tr *Transactor) Conn(ctx context.Context) Connection { //nolint:ireturn
	if tx, ok := ctx.Value(txKey{}).(*sql.Tx); ok {
		return tx
	}
	return tr.database
}

func NewTransactor(database database) *Transactor {
	return &Transactor{
		database: database,
	}
}
