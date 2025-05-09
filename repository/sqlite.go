package repository

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// SQLiteRepository implements Repository using SQLite
type SQLiteRepository struct {
	db     *sql.DB
	dbPath string
}

// NewSQLiteRepository creates a new SQLite repository instance
func NewSQLiteRepository() Repository {
	// Default to data directory in user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	// Create data directory if it doesn't exist
	dataDir := filepath.Join(homeDir, ".theary")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		// Fallback to current directory if home directory is not accessible
		dataDir = "."
	}

	return &SQLiteRepository{
		dbPath: filepath.Join(dataDir, "theary.db"),
	}
}

// Initialize sets up the SQLite database
func (r *SQLiteRepository) Initialize(ctx context.Context) error {
	// Open SQLite database
	db, err := sql.Open("sqlite3", r.dbPath)
	if err != nil {
		return err
	}

	// Create nodes table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS nodes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			label TEXT NOT NULL,
			parent_id INTEGER,
			FOREIGN KEY (parent_id) REFERENCES nodes(id)
		)
	`)
	if err != nil {
		db.Close()
		return err
	}

	r.db = db
	return nil
}

// Cleanup closes the database connection
func (r *SQLiteRepository) Cleanup(ctx context.Context) error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}

// CreateNode creates a new node in the database
func (r *SQLiteRepository) CreateNode(ctx context.Context, label string, parentID *int64) (int64, error) {
	// Check if parent exists
	if parentID != nil {
		var exists bool
		err := r.db.QueryRow("SELECT EXISTS(SELECT 1 FROM nodes WHERE id = ?)", *parentID).Scan(&exists)
		if err != nil {
			return 0, err
		}
		if !exists {
			return 0, ErrNodeNotFound
		}
	}

	result, err := r.db.Exec("INSERT INTO nodes (label, parent_id) VALUES (?, ?)", label, parentID)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// GetNode retrieves a node by ID
func (r *SQLiteRepository) GetNode(ctx context.Context, id int64) (*Node, error) {
	var node Node
	var parentID sql.NullInt64
	err := r.db.QueryRow("SELECT id, label, parent_id FROM nodes WHERE id = ?", id).
		Scan(&node.ID, &node.Label, &parentID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNodeNotFound
		}
		return nil, err
	}
	if parentID.Valid {
		node.ParentID = &parentID.Int64
	}
	return &node, nil
}

// GetAllNodes retrieves all nodes from the database
func (r *SQLiteRepository) GetAllNodes(ctx context.Context) ([]*Node, error) {
	rows, err := r.db.Query("SELECT id, label, parent_id FROM nodes")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []*Node
	for rows.Next() {
		var node Node
		var parentID sql.NullInt64
		if err := rows.Scan(&node.ID, &node.Label, &parentID); err != nil {
			return nil, err
		}
		if parentID.Valid {
			node.ParentID = &parentID.Int64
		}
		nodes = append(nodes, &node)
	}
	return nodes, rows.Err()
}

// UpdateNode updates a node's properties
func (r *SQLiteRepository) UpdateNode(ctx context.Context, id int64, label string, parentID *int64) error {
	result, err := r.db.Exec("UPDATE nodes SET label = ?, parent_id = ? WHERE id = ?", label, parentID, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNodeNotFound
	}
	return nil
}

// DeleteNode deletes a node and its children
func (r *SQLiteRepository) DeleteNode(ctx context.Context, id int64) error {
	// First, delete all child nodes recursively
	rows, err := r.db.Query("SELECT id FROM nodes WHERE parent_id = ?", id)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var childID int64
		if err := rows.Scan(&childID); err != nil {
			return err
		}
		if err := r.DeleteNode(ctx, childID); err != nil {
			return err
		}
	}

	// Then delete the node itself
	result, err := r.db.Exec("DELETE FROM nodes WHERE id = ?", id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNodeNotFound
	}
	return nil
}
