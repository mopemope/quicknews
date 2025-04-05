package database

import (
	"context"
	"log/slog"
	"sync"

	"github.com/cockroachdb/errors"
	"github.com/mopemope/quicknews/ent"
)

var databaseMutex = &sync.Mutex{}

func WithTx(ctx context.Context, client *ent.Client, fn func(tx *ent.Tx) error) error {
	databaseMutex.Lock()
	defer databaseMutex.Unlock()

	tx, err := client.Tx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if v := recover(); v != nil {
			if err := tx.Rollback(); err != nil {
				slog.Error("failed to rollback transaction", "error", err)
			}
		}
	}()

	if err := fn(tx); err != nil {
		if rerr := tx.Rollback(); rerr != nil {
			return errors.Wrapf(rerr, "%w: rolling back transaction", err)
		}
		return err
	}
	if err := tx.Commit(); err != nil {
		return errors.Wrapf(err, "committing transaction")
	}
	return nil
}
