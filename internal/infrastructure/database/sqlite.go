package database

import (
	"database/sql"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	_ "modernc.org/sqlite"
)

// Open creates a new SQLite connection with WAL mode and foreign keys enabled.
func Open(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath+"?_pragma=journal_mode(wal)&_pragma=foreign_keys(on)")
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	if err := runMigrations(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("running migrations: %w", err)
	}

	return db, nil
}

func runMigrations(db *sql.DB) error {
	// Create migration tracking table if it doesn't exist.
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		name TEXT PRIMARY KEY,
		applied_at DATETIME NOT NULL DEFAULT (datetime('now'))
	)`); err != nil {
		return fmt.Errorf("creating schema_migrations table: %w", err)
	}

	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("reading migrations dir: %w", err)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	// Backfill: if the DB already has tables but no migration records
	// (upgrading from before migration tracking), mark existing migrations
	// as applied by checking whether their effects are already present.
	var tracked int
	if err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&tracked); err != nil {
		return fmt.Errorf("counting tracked migrations: %w", err)
	}
	if tracked == 0 {
		var hasRecipesTable int
		db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='recipes'").Scan(&hasRecipesTable)
		if hasRecipesTable > 0 {
			// DB was created before migration tracking — seed all migrations
			// whose effects are already present.
			for _, entry := range entries {
				if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
					continue
				}
				if migrationAlreadyApplied(db, entry.Name()) {
					db.Exec("INSERT OR IGNORE INTO schema_migrations (name) VALUES (?)", entry.Name())
				}
			}
		}
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		// Skip migrations that have already been applied.
		var count int
		if err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE name = ?", entry.Name()).Scan(&count); err != nil {
			return fmt.Errorf("checking migration %s: %w", entry.Name(), err)
		}
		if count > 0 {
			continue
		}

		data, err := migrationsFS.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return fmt.Errorf("reading migration %s: %w", entry.Name(), err)
		}
		if _, err := db.Exec(string(data)); err != nil {
			return fmt.Errorf("executing migration %s: %w", entry.Name(), err)
		}

		if _, err := db.Exec("INSERT INTO schema_migrations (name) VALUES (?)", entry.Name()); err != nil {
			return fmt.Errorf("recording migration %s: %w", entry.Name(), err)
		}
	}

	return nil
}

// migrationAlreadyApplied checks whether a migration's effects are already
// present in the database. Used for backfilling the tracking table on
// databases created before migration tracking was added.
func migrationAlreadyApplied(db *sql.DB, name string) bool {
	switch name {
	case "001_init.sql":
		var count int
		db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='recipes'").Scan(&count)
		return count > 0
	case "002_add_notes.sql":
		var count int
		db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('recipes') WHERE name='notes'").Scan(&count)
		return count > 0
	case "003_add_users.sql":
		var count int
		db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='users'").Scan(&count)
		return count > 0
	case "005_add_image_path.sql":
		var count int
		db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('recipes') WHERE name='image_path'").Scan(&count)
		return count > 0
	default:
		return false
	}
}
