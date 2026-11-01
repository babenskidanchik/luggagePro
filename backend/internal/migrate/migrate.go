package migrate

import (
	"context"
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"

	"github.com/babenskidanchik/luggage-platform/backend/migrations"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Run(ctx context.Context, db *pgxpool.Pool) error {
	if err := ensureMigrationsTable(ctx, db); err != nil {
		return err
	}

	files, err := migrationFiles()
	if err != nil {
		return err
	}

	for _, name := range files {
		version, err := migrationVersion(name)
		if err != nil {
			return err
		}

		applied, err := isApplied(ctx, db, version)
		if err != nil {
			return err
		}

		if applied {
			continue
		}

		if err := apply(ctx, db, version, name); err != nil {
			return err
		}

		fmt.Printf("applied migration %s\n", name)
	}

	return nil
}

func ensureMigrationsTable(
	ctx context.Context,
	db *pgxpool.Pool,
) error {
	const query = `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`

	if _, err := db.Exec(ctx, query); err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}

	return nil
}

func migrationFiles() ([]string, error) {
	entries, err := fs.ReadDir(migrations.FS, ".")
	if err != nil {
		return nil, fmt.Errorf("read migrations: %w", err)
	}

	var files []string

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()

		if !strings.HasSuffix(name, ".sql") {
			continue
		}

		files = append(files, name)
	}

	sort.Strings(files)

	return files, nil
}

func migrationVersion(name string) (int, error) {
	parts := strings.SplitN(name, "_", 2)

	if len(parts) != 2 {
		return 0, fmt.Errorf(
			"invalid migration filename %q",
			name,
		)
	}

	version, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, fmt.Errorf(
			"parse migration version %q: %w",
			name,
			err,
		)
	}

	return version, nil
}

func isApplied(
	ctx context.Context,
	db *pgxpool.Pool,
	version int,
) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT 1
			FROM schema_migrations
			WHERE version = $1
		)
	`

	var applied bool

	if err := db.QueryRow(
		ctx,
		query,
		version,
	).Scan(&applied); err != nil {
		return false, fmt.Errorf(
			"check migration %d: %w",
			version,
			err,
		)
	}

	return applied, nil
}

func apply(
	ctx context.Context,
	db *pgxpool.Pool,
	version int,
	name string,
) error {
	sqlBytes, err := migrations.FS.ReadFile(name)
	if err != nil {
		return fmt.Errorf(
			"read migration %s: %w",
			name,
			err,
		)
	}

	tx, err := db.Begin(ctx)
	if err != nil {
		return fmt.Errorf(
			"begin migration %s: %w",
			name,
			err,
		)
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if _, err := tx.Exec(ctx, string(sqlBytes)); err != nil {
		return fmt.Errorf(
			"execute migration %s: %w",
			name,
			err,
		)
	}

	const insertQuery = `
		INSERT INTO schema_migrations (
			version,
			name
		)
		VALUES ($1, $2)
	`

	if _, err := tx.Exec(
		ctx,
		insertQuery,
		version,
		name,
	); err != nil {
		return fmt.Errorf(
			"record migration %s: %w",
			name,
			err,
		)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf(
			"commit migration %s: %w",
			name,
			err,
		)
	}

	return nil
}
