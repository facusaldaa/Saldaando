package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// DB wraps the database connection
type DB struct {
	conn *sql.DB
}

// NewDB creates a new database connection
func NewDB(dbPath string) (*DB, error) {
	// Create data directory if it doesn't exist
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	conn, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=1")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db := &DB{conn: conn}

	// Run migrations
	if err := db.migrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

// GetConn returns the underlying sql.DB connection
func (db *DB) GetConn() *sql.DB {
	return db.conn
}

// migrate runs database migrations
func (db *DB) migrate() error {
	conn := db.conn

	// Create tables if they don't exist (without new columns first)
	schemaSQL := `
	-- Users table
	CREATE TABLE IF NOT EXISTS users (
		telegram_id INTEGER PRIMARY KEY,
		username TEXT,
		display_name TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- Lobbies (couples) - initial schema without new columns
	CREATE TABLE IF NOT EXISTS lobbies (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user1_telegram_id INTEGER,
		user2_telegram_id INTEGER,
		account_type TEXT CHECK(account_type IN ('separate', 'shared')),
		user1_salary_percentage REAL DEFAULT 0.5,
		user2_salary_percentage REAL DEFAULT 0.5,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user1_telegram_id) REFERENCES users(telegram_id),
		FOREIGN KEY (user2_telegram_id) REFERENCES users(telegram_id)
	);

	-- Categories (predefined + custom)
	CREATE TABLE IF NOT EXISTS categories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		lobby_id INTEGER,
		name TEXT NOT NULL,
		is_default BOOLEAN DEFAULT 0,
		FOREIGN KEY (lobby_id) REFERENCES lobbies(id)
	);

	-- Payment Methods (credit cards, cash, etc.)
	CREATE TABLE IF NOT EXISTS payment_methods (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		lobby_id INTEGER,
		name TEXT NOT NULL,
		type TEXT CHECK(type IN ('credit_card', 'debit_card', 'cash', 'bank_transfer', 'other')),
		owner_telegram_id INTEGER,
		closing_day INTEGER,
		billing_cycle_days INTEGER DEFAULT 30,
		is_active BOOLEAN DEFAULT 1,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (lobby_id) REFERENCES lobbies(id),
		FOREIGN KEY (owner_telegram_id) REFERENCES users(telegram_id)
	);

	-- Expenses
	CREATE TABLE IF NOT EXISTS expenses (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		lobby_id INTEGER,
		spender_telegram_id INTEGER,
		payment_method_id INTEGER,
		amount REAL NOT NULL,
		description TEXT,
		category TEXT,
		expense_date DATE NOT NULL,
		billing_period_start DATE,
		billing_period_end DATE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (lobby_id) REFERENCES lobbies(id),
		FOREIGN KEY (spender_telegram_id) REFERENCES users(telegram_id),
		FOREIGN KEY (payment_method_id) REFERENCES payment_methods(id)
	);
	`

	if _, err := conn.Exec(schemaSQL); err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}

	// Migrate existing tables: add new columns if they don't exist
	if err := db.migrateAddColumns(); err != nil {
		return fmt.Errorf("failed to migrate columns: %w", err)
	}

	// Create indexes (after ensuring columns exist)
	indexSQL := `
	CREATE INDEX IF NOT EXISTS idx_expenses_lobby_date ON expenses(lobby_id, expense_date);
	CREATE INDEX IF NOT EXISTS idx_expenses_billing_period ON expenses(billing_period_start, billing_period_end);
	CREATE INDEX IF NOT EXISTS idx_expenses_payment_method ON expenses(payment_method_id);
	CREATE INDEX IF NOT EXISTS idx_payment_methods_lobby ON payment_methods(lobby_id, is_active);
	`

	if _, err := conn.Exec(indexSQL); err != nil {
		// Some indexes might fail if columns don't exist, that's ok
	}

	return nil
}

// migrateAddColumns adds new columns to existing tables
func (db *DB) migrateAddColumns() error {
	conn := db.conn

	// Add language column to users table if it doesn't exist
	var userLangCount int
	err := conn.QueryRow(`
		SELECT COUNT(*) FROM pragma_table_info('users') WHERE name='language'
	`).Scan(&userLangCount)
	if err != nil {
		// Table might not exist yet, that's ok
		return nil
	}
	if userLangCount == 0 {
		_, err = conn.Exec(`ALTER TABLE users ADD COLUMN language TEXT DEFAULT 'en'`)
		if err != nil {
			// Column might already exist, ignore error
			fmt.Printf("Warning: Could not add language column: %v\n", err)
		}
	}

	// Add invite_token column to lobbies table if it doesn't exist
	var inviteTokenCount int
	err = conn.QueryRow(`
		SELECT COUNT(*) FROM pragma_table_info('lobbies') WHERE name='invite_token'
	`).Scan(&inviteTokenCount)
	if err != nil {
		// Table might not exist yet, that's ok
		return nil
	}
	if inviteTokenCount == 0 {
		_, err = conn.Exec(`ALTER TABLE lobbies ADD COLUMN invite_token TEXT`)
		if err != nil {
			// Log but don't fail - column might already exist
			fmt.Printf("Warning: Could not add invite_token column: %v\n", err)
		} else {
			// Column was added successfully
			inviteTokenCount = 1
		}
	}

	// Add group_chat_id column to lobbies table if it doesn't exist
	var groupChatIDCount int
	err = conn.QueryRow(`
		SELECT COUNT(*) FROM pragma_table_info('lobbies') WHERE name='group_chat_id'
	`).Scan(&groupChatIDCount)
	if err != nil {
		// Table might not exist yet, that's ok
		return nil
	}
	if groupChatIDCount == 0 {
		_, err = conn.Exec(`ALTER TABLE lobbies ADD COLUMN group_chat_id INTEGER`)
		if err != nil {
			// Log but don't fail - column might already exist
			fmt.Printf("Warning: Could not add group_chat_id column: %v\n", err)
		}
	}

	// Create index on invite_token if column exists
	if inviteTokenCount > 0 {
		_, _ = conn.Exec(`CREATE INDEX IF NOT EXISTS idx_lobbies_invite_token ON lobbies(invite_token)`)
	}

	return nil
}
