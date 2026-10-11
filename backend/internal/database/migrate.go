package database

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"log/slog"
	"sort"
	"strconv"
	"strings"
)

// Migration files are named NNNNNN_description.up.sql / .down.sql, the layout
// golang-migrate uses, so the CLI stays an option if this runner is outgrown.
type migration struct {
	version int64
	name    string
	up      string
	down    string
}

const createMigrationsTable = `
CREATE TABLE IF NOT EXISTS schema_migrations (
    version    BIGINT   NOT NULL,
    name       VARCHAR(191) NOT NULL,
    applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (version)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

// MigrateUp applies every migration that is not recorded yet, oldest first.
func MigrateUp(ctx context.Context, db *sql.DB, fsys fs.FS) error {
	all, applied, err := load(ctx, db, fsys)
	if err != nil {
		return err
	}

	count := 0
	for _, m := range all {
		if applied[m.version] {
			continue
		}
		if err := exec(ctx, db, m.up); err != nil {
			return fmt.Errorf("apply %d_%s: %w", m.version, m.name, err)
		}
		if _, err := db.ExecContext(ctx,
			`INSERT INTO schema_migrations (version, name) VALUES (?, ?)`, m.version, m.name); err != nil {
			return err
		}
		slog.Info("migration applied", "version", m.version, "name", m.name)
		count++
	}
	if count == 0 {
		slog.Info("migrations already up to date")
	}
	return nil
}

// MigrateDown rolls back the newest applied migrations, newest first.
func MigrateDown(ctx context.Context, db *sql.DB, fsys fs.FS, steps int) error {
	all, applied, err := load(ctx, db, fsys)
	if err != nil {
		return err
	}

	for i := len(all) - 1; i >= 0 && steps > 0; i-- {
		m := all[i]
		if !applied[m.version] {
			continue
		}
		if strings.TrimSpace(m.down) == "" {
			return fmt.Errorf("migration %d_%s has no down file", m.version, m.name)
		}
		if err := exec(ctx, db, m.down); err != nil {
			return fmt.Errorf("roll back %d_%s: %w", m.version, m.name, err)
		}
		if _, err := db.ExecContext(ctx, `DELETE FROM schema_migrations WHERE version = ?`, m.version); err != nil {
			return err
		}
		slog.Info("migration rolled back", "version", m.version, "name", m.name)
		steps--
	}
	return nil
}

// MigrationStatus reports the highest applied version and how many are pending.
func MigrationStatus(ctx context.Context, db *sql.DB, fsys fs.FS) (current int64, pending int, err error) {
	all, applied, err := load(ctx, db, fsys)
	if err != nil {
		return 0, 0, err
	}
	for _, m := range all {
		if applied[m.version] {
			current = max(current, m.version)
		} else {
			pending++
		}
	}
	return current, pending, nil
}

func load(ctx context.Context, db *sql.DB, fsys fs.FS) ([]migration, map[int64]bool, error) {
	if _, err := db.ExecContext(ctx, createMigrationsTable); err != nil {
		return nil, nil, err
	}

	all, err := read(fsys)
	if err != nil {
		return nil, nil, err
	}

	rows, err := db.QueryContext(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	applied := map[int64]bool{}
	for rows.Next() {
		var v int64
		if err := rows.Scan(&v); err != nil {
			return nil, nil, err
		}
		applied[v] = true
	}
	return all, applied, rows.Err()
}

func read(fsys fs.FS) ([]migration, error) {
	entries, err := fs.Glob(fsys, "*.sql")
	if err != nil {
		return nil, err
	}

	byVersion := map[int64]*migration{}
	for _, entry := range entries {
		version, name, direction, err := parseName(entry)
		if err != nil {
			return nil, err
		}
		body, err := fs.ReadFile(fsys, entry)
		if err != nil {
			return nil, err
		}
		m := byVersion[version]
		if m == nil {
			m = &migration{version: version, name: name}
			byVersion[version] = m
		}
		if direction == "up" {
			m.up = string(body)
		} else {
			m.down = string(body)
		}
	}

	all := make([]migration, 0, len(byVersion))
	for _, m := range byVersion {
		if strings.TrimSpace(m.up) == "" {
			return nil, fmt.Errorf("migration %d_%s has no up file", m.version, m.name)
		}
		all = append(all, *m)
	}
	sort.Slice(all, func(i, j int) bool { return all[i].version < all[j].version })
	return all, nil
}

func parseName(filename string) (version int64, name, direction string, err error) {
	base := strings.TrimSuffix(filename, ".sql")
	dot := strings.LastIndex(base, ".")
	if dot < 0 {
		return 0, "", "", fmt.Errorf("migration %q: expected NNNNNN_name.up.sql", filename)
	}
	direction = base[dot+1:]
	if direction != "up" && direction != "down" {
		return 0, "", "", fmt.Errorf("migration %q: direction must be up or down", filename)
	}

	rest := base[:dot]
	underscore := strings.Index(rest, "_")
	if underscore < 0 {
		return 0, "", "", fmt.Errorf("migration %q: expected NNNNNN_name.up.sql", filename)
	}
	version, err = strconv.ParseInt(rest[:underscore], 10, 64)
	if err != nil {
		return 0, "", "", fmt.Errorf("migration %q: bad version: %w", filename, err)
	}
	return version, rest[underscore+1:], direction, nil
}

// exec runs a migration file statement by statement, stripping comments and empty lines.
// If an ALTER TABLE statement fails because a column or constraint already exists
// (e.g. from a partially applied migration in MySQL/MariaDB which does not roll back DDL),
// it ignores the duplicate error so the migration can complete safely.
func exec(ctx context.Context, db *sql.DB, body string) error {
	for _, statement := range strings.Split(body, ";") {
		clean := cleanStatement(statement)
		if clean == "" {
			continue
		}
		if _, err := db.ExecContext(ctx, clean); err != nil {
			if isIgnorableDDLError(err) {
				slog.Warn("ignoring idempotent DDL error", "error", err)
				continue
			}
			return err
		}
	}
	return nil
}

func isIgnorableDDLError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	// MySQL/MariaDB Error 1060: Duplicate column name
	// MySQL/MariaDB Error 1061: Duplicate key name
	// MySQL/MariaDB Error 1050: Table already exists
	// MySQL/MariaDB Error 1826: Duplicate foreign key constraint name
	return strings.Contains(msg, "duplicate column name") ||
		strings.Contains(msg, "duplicate key name") ||
		strings.Contains(msg, "already exists") ||
		strings.Contains(msg, "duplicate foreign key")
}

func cleanStatement(raw string) string {
	var lines []string
	for _, line := range strings.Split(raw, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "--") || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if trimmed != "" {
			lines = append(lines, line)
		}
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}
