// Package migrations applies the explicit, transactional Cully schema migration.
package migrations

import (
	"context"
	_ "embed"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed 002_source_only.sql
var schema string

//go:embed 003_session_ref.sql
var sessionRefSchema string

//go:embed 004_sessions_tasks.sql
var sessionsTasksSchema string

func Apply(ctx context.Context, pool *pgxpool.Pool) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock(724812093)"); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, "CREATE TABLE IF NOT EXISTS cully_schema_versions(version integer PRIMARY KEY, applied_at timestamptz NOT NULL DEFAULT now())"); err != nil {
		return err
	}
	for _, m := range []struct {
		version int
		sql     string
	}{{2, schema}, {3, sessionRefSchema}, {4, sessionsTasksSchema}} {
		var exists bool
		if err = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM cully_schema_versions WHERE version=$1)", m.version).Scan(&exists); err != nil {
			return err
		}
		if exists {
			continue
		}
		if _, err = tx.Exec(ctx, m.sql); err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, "INSERT INTO cully_schema_versions(version) VALUES($1)", m.version); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
