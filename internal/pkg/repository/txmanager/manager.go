package txmanager

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type (
	// TxManager for spawning new txs.
	TxManager interface {
		Do(ctx context.Context, starter TxStarter, fn TxFunc) error
	}

	// TxStarter must implement new transaction spawn.
	TxStarter interface {
		BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
	}

	TxFunc func(t pgx.Tx) error
)

type Manager struct{}

// New transaction manager.
func New() *Manager {
	return &Manager{}
}

// Do execs given fn in transaction and commits or rollbacks it.
func (m *Manager) Do(ctx context.Context, db TxStarter, fn TxFunc) (err error) {
	// todo add configurable transaction opts
	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("failed begin pg tx: %w", err)
	}

	defer func() {
		// it's safe to rollback committed tx
		_ = tx.Rollback(ctx)
	}()

	err = fn(tx)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}
