package migrate

import (
	"context"
	"embed"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed sql/*.sql
var migrationFS embed.FS

// Run executes all pending up-migrations against the database.
// It is safe to call on every startup — already-applied migrations are skipped.
func Run(ctx context.Context, pool *pgxpool.Pool) error {
	// 1. Ensure the tracking table exists
	if _, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INT PRIMARY KEY,
			name    TEXT NOT NULL,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`); err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}

	// 2. Read embedded migration files
	entries, err := migrationFS.ReadDir("sql")
	if err != nil {
		return fmt.Errorf("read embedded migrations: %w", err)
	}

	// Collect .up.sql files, sorted by name (which gives version order)
	var upFiles []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".up.sql") {
			upFiles = append(upFiles, e.Name())
		}
	}
	sort.Strings(upFiles)

	if len(upFiles) == 0 {
		log.Println("[migrate] No migration files found")
		return nil
	}

	// 3. Get already-applied versions
	applied := make(map[int]bool)
	rows, err := pool.Query(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return fmt.Errorf("query applied migrations: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			return err
		}
		applied[v] = true
	}
	if err := rows.Err(); err != nil {
		return err
	}

	// 4. Apply pending migrations
	for _, filename := range upFiles {
		version, name, err := parseFilename(filename)
		if err != nil {
			return fmt.Errorf("parse migration filename %q: %w", filename, err)
		}

		if applied[version] {
			continue
		}

		sql, err := migrationFS.ReadFile("sql/" + filename)
		if err != nil {
			return fmt.Errorf("read migration %q: %w", filename, err)
		}

		log.Printf("[migrate] Applying %03d_%s ...", version, name)

		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin tx for migration %d: %w", version, err)
		}

		if _, err := tx.Exec(ctx, string(sql)); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("execute migration %d: %w", version, err)
		}

		if _, err := tx.Exec(ctx,
			`INSERT INTO schema_migrations (version, name) VALUES ($1, $2)`,
			version, name,
		); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("record migration %d: %w", version, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit migration %d: %w", version, err)
		}

		log.Printf("[migrate] Applied %03d_%s", version, name)
	}

	log.Printf("[migrate] All migrations up to date (%d files)", len(upFiles))
	return nil
}

// parseFilename extracts version number and name from "001_create_sessions.up.sql"
func parseFilename(filename string) (int, string, error) {
	// Remove ".up.sql" suffix
	base := strings.TrimSuffix(filename, ".up.sql")
	// Split on first underscore: "001" + "create_sessions"
	parts := strings.SplitN(base, "_", 2)
	if len(parts) != 2 {
		return 0, "", fmt.Errorf("unexpected format: %s", filename)
	}
	var version int
	if _, err := fmt.Sscanf(parts[0], "%d", &version); err != nil {
		return 0, "", fmt.Errorf("parse version from %q: %w", parts[0], err)
	}
	return version, parts[1], nil
}
