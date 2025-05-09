package migrations

import (
	"context"
	"database/sql"
	"fmt"
)

// Migration represents a database migration
type Migration struct {
	ID   int
	Name string
	Up   string
	Down string
}

// Migrations is a list of all database migrations
var Migrations = []Migration{
	{
		ID:   1,
		Name: "create_nodes_table",
		Up: `
			CREATE TABLE IF NOT EXISTS nodes (
				id SERIAL PRIMARY KEY,
				label TEXT NOT NULL,
				parent_id INTEGER REFERENCES nodes(id),
				created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
			)
		`,
		Down: `DROP TABLE IF EXISTS nodes`,
	},
	{
		ID:   2,
		Name: "create_updated_at_trigger",
		Up: `
			CREATE OR REPLACE FUNCTION update_updated_at_column()
			RETURNS TRIGGER AS $$
			BEGIN
				NEW.updated_at = CURRENT_TIMESTAMP;
				RETURN NEW;
			END;
			$$ language 'plpgsql';

			DROP TRIGGER IF EXISTS update_nodes_updated_at ON nodes;
			CREATE TRIGGER update_nodes_updated_at
				BEFORE UPDATE ON nodes
				FOR EACH ROW
				EXECUTE FUNCTION update_updated_at_column();
		`,
		Down: `
			DROP TRIGGER IF EXISTS update_nodes_updated_at ON nodes;
			DROP FUNCTION IF EXISTS update_updated_at_column();
		`,
	},
}

// RunMigrations executes all pending migrations
func RunMigrations(ctx context.Context, db *sql.DB) error {
	// Create migrations table if it doesn't exist
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS migrations (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("error creating migrations table: %w", err)
	}

	// Get applied migrations
	rows, err := db.QueryContext(ctx, "SELECT id FROM migrations ORDER BY id")
	if err != nil {
		return fmt.Errorf("error querying applied migrations: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			fmt.Printf("Warning: Error closing rows: %v\n", err)
		}
	}()

	applied := make(map[int]bool)
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return fmt.Errorf("error scanning migration id: %w", err)
		}
		applied[id] = true
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating migrations: %w", err)
	}

	// Begin transaction
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error beginning transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			fmt.Printf("Warning: Error rolling back transaction: %v\n", err)
		}
	}()

	// Apply pending migrations
	for _, migration := range Migrations {
		if !applied[migration.ID] {
			// Execute migration
			if _, err := tx.ExecContext(ctx, migration.Up); err != nil {
				return fmt.Errorf("error executing migration %d (%s): %w", migration.ID, migration.Name, err)
			}

			// Record migration
			if _, err := tx.ExecContext(ctx, "INSERT INTO migrations (id, name) VALUES ($1, $2)",
				migration.ID, migration.Name); err != nil {
				return fmt.Errorf("error recording migration %d (%s): %w", migration.ID, migration.Name, err)
			}
		}
	}

	return tx.Commit()
}

// RollbackMigration rolls back the last applied migration
func RollbackMigration(ctx context.Context, db *sql.DB) error {
	// Get the last applied migration
	var lastMigration Migration
	err := db.QueryRowContext(ctx, `
		SELECT m.id, m.name
		FROM migrations m
		ORDER BY m.id DESC
		LIMIT 1
	`).Scan(&lastMigration.ID, &lastMigration.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("no migrations to rollback")
		}
		return fmt.Errorf("error querying last migration: %w", err)
	}

	// Find the migration in our list
	var migration Migration
	for _, m := range Migrations {
		if m.ID == lastMigration.ID {
			migration = m
			break
		}
	}

	// Begin transaction
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error beginning transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			fmt.Printf("Warning: Error rolling back transaction: %v\n", err)
		}
	}()

	// Execute rollback
	if _, err := tx.ExecContext(ctx, migration.Down); err != nil {
		return fmt.Errorf("error rolling back migration %d (%s): %w", migration.ID, migration.Name, err)
	}

	// Remove migration record
	if _, err := tx.ExecContext(ctx, "DELETE FROM migrations WHERE id = $1", migration.ID); err != nil {
		return fmt.Errorf("error removing migration record %d (%s): %w", migration.ID, migration.Name, err)
	}

	return tx.Commit()
}
