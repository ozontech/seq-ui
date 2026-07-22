package txmanager

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type (
	// TxStarter must implement new transaction spawn.
	TxStarter interface {
		Begin(ctx context.Context) (pgx.Tx, error)
		BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
	}

	TxFunc func(t pgx.Tx) error
)

type Manager struct {
	db TxStarter
}

// New transaction manager.
func New(db TxStarter) *Manager {
	return &Manager{db: db}
}

// DoTx executes fn in a transaction with the given options.
func (m *Manager) DoTx(ctx context.Context, txOptions pgx.TxOptions, fn TxFunc) error {
	tx, err := m.db.BeginTx(ctx, txOptions)
	if err != nil {
		return fmt.Errorf("failed begin pg tx: %w", err)
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// Do executes fn in a transaction with default options.
func (m *Manager) Do(ctx context.Context, fn TxFunc) error {
	return m.DoTx(ctx, pgx.TxOptions{}, fn)
}
