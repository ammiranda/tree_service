package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/ammiranda/tree_service/config"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

// PostgresRepository implements Repository using PostgreSQL
type PostgresRepository struct {
	db     *sql.DB
	config *config.DatabaseConfig
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(cfgProvider config.Provider) (*PostgresRepository, error) {
	ctx := context.Background()
	cfg, err := config.GetDatabaseConfig(ctx, cfgProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to get database config: %w", err)
	}

	return &PostgresRepository{
		config: cfg,
	}, nil
}

// Initialize sets up the PostgreSQL database
func (r *PostgresRepository) Initialize(ctx context.Context) error {
	// Construct connection string using standard format
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		r.config.User,
		r.config.Password,
		r.config.Host,
		r.config.Port,
		r.config.DBName,
		r.config.SSLMode,
	)

	fmt.Printf("Attempting to connect to database at %s:%d\n", r.config.Host, r.config.Port)

	// Open database connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("error connecting to database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	fmt.Println("Testing database connection...")

	// Test the connection
	if err := db.PingContext(ctx); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			fmt.Printf("Warning: Error closing database connection: %v\n", closeErr)
		}
		return fmt.Errorf("error pinging database: %w", err)
	}

	fmt.Println("Database connection successful, running migrations...")

	// Run migrations
	if err := r.runMigrations(db); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			fmt.Printf("Warning: Error closing database connection: %v\n", closeErr)
		}
		return fmt.Errorf("error running migrations: %w", err)
	}

	fmt.Println("Migrations completed successfully")

	r.db = db
	return nil
}

// runMigrations executes database migrations
func (r *PostgresRepository) runMigrations(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("error creating migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file:///app/migrations",
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("error creating migration instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("error running migrations: %w", err)
	}

	return nil
}

// Cleanup closes the database connection
func (r *PostgresRepository) Cleanup(ctx context.Context) error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}

// CreateNode creates a new node in the database
func (r *PostgresRepository) CreateNode(ctx context.Context, label string, parentID *int64) (int64, error) {
	if label == "" {
		return 0, ErrInvalidInput
	}

	// Check if parent exists
	if parentID != nil {
		exists, err := r.nodeExists(ctx, *parentID)
		if err != nil {
			return 0, err
		}
		if !exists {
			return 0, ErrNodeNotFound
		}
	}

	var id int64
	err := r.db.QueryRowContext(ctx,
		"INSERT INTO nodes (label, parent_id) VALUES ($1, $2) RETURNING id",
		label, parentID,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("error creating node: %w", err)
	}
	return id, nil
}

// GetNode retrieves a node by ID
func (r *PostgresRepository) GetNode(ctx context.Context, id int64) (*Node, error) {
	var node Node
	var parentID sql.NullInt64
	err := r.db.QueryRowContext(ctx,
		"SELECT id, label, parent_id FROM nodes WHERE id = $1",
		id,
	).Scan(&node.ID, &node.Label, &parentID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNodeNotFound
		}
		return nil, fmt.Errorf("error getting node: %w", err)
	}
	if parentID.Valid {
		node.ParentID = &parentID.Int64
	}
	return &node, nil
}

// GetAllNodes retrieves all nodes from the database with pagination
func (r *PostgresRepository) GetAllNodes(ctx context.Context, page int, pageSize int) ([]*Node, int64, error) {
	// Get total count
	var total int64
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM nodes").Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("error getting total count: %w", err)
	}

	// Calculate offset
	offset := (page - 1) * pageSize

	// Get paginated nodes
	rows, err := r.db.QueryContext(ctx,
		"SELECT id, label, parent_id FROM nodes ORDER BY id LIMIT $1 OFFSET $2",
		pageSize, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("error getting nodes: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			fmt.Printf("Warning: Error closing rows: %v\n", err)
		}
	}()

	var nodes []*Node
	for rows.Next() {
		var node Node
		var parentID sql.NullInt64
		if err := rows.Scan(&node.ID, &node.Label, &parentID); err != nil {
			return nil, 0, fmt.Errorf("error scanning node: %w", err)
		}
		if parentID.Valid {
			node.ParentID = &parentID.Int64
		}
		nodes = append(nodes, &node)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating nodes: %w", err)
	}

	return nodes, total, nil
}

// UpdateNode updates a node's properties
func (r *PostgresRepository) UpdateNode(ctx context.Context, id int64, label string, parentID *int64) error {
	if label == "" {
		return ErrInvalidInput
	}

	// Check if node exists
	exists, err := r.nodeExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return ErrNodeNotFound
	}

	// Check if new parent exists
	if parentID != nil {
		exists, err := r.nodeExists(ctx, *parentID)
		if err != nil {
			return err
		}
		if !exists {
			return ErrNodeNotFound
		}
	}

	result, err := r.db.ExecContext(ctx,
		"UPDATE nodes SET label = $1, parent_id = $2 WHERE id = $3",
		label, parentID, id,
	)
	if err != nil {
		return fmt.Errorf("error updating node: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}
	if rows == 0 {
		return ErrNodeNotFound
	}
	return nil
}

// DeleteNode deletes a node and its children
func (r *PostgresRepository) DeleteNode(ctx context.Context, id int64) error {
	// Use a transaction to ensure atomicity
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error beginning transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			// Log the error but don't return it since we're in a defer
			fmt.Printf("Error rolling back transaction: %v\n", err)
		}
	}()

	// Delete all child nodes recursively using a CTE
	_, err = tx.ExecContext(ctx, `
		WITH RECURSIVE children AS (
			SELECT id FROM nodes WHERE parent_id = $1
			UNION ALL
			SELECT n.id FROM nodes n
			INNER JOIN children c ON n.parent_id = c.id
		)
		DELETE FROM nodes WHERE id IN (SELECT id FROM children)
	`, id)
	if err != nil {
		return fmt.Errorf("error deleting child nodes: %w", err)
	}

	// Delete the node itself
	result, err := tx.ExecContext(ctx, "DELETE FROM nodes WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("error deleting node: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ErrNodeNotFound
	}

	return tx.Commit()
}

// nodeExists checks if a node exists
func (r *PostgresRepository) nodeExists(ctx context.Context, id int64) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx,
		"SELECT EXISTS(SELECT 1 FROM nodes WHERE id = $1)",
		id,
	).Scan(&exists)
	return exists, err
}
