package service

import (
	"botGastosPareja/internal/database"
	"botGastosPareja/pkg/i18n"
	"database/sql"
	"fmt"
	"time"
)

// UserService handles user-related operations
type UserService struct {
	db *database.DB
}

// NewUserService creates a new user service
func NewUserService(db *database.DB) *UserService {
	return &UserService{db: db}
}

// GetOrCreateUser gets an existing user or creates a new one
func (s *UserService) GetOrCreateUser(telegramID int64, username string, displayName string) (*database.User, error) {
	conn := s.db.GetConn()

	// Try to get existing user
	var user database.User
	query := `SELECT telegram_id, username, display_name, language, created_at 
	          FROM users WHERE telegram_id = ?`

	err := conn.QueryRow(query, telegramID).Scan(
		&user.TelegramID,
		&user.Username,
		&user.DisplayName,
		&user.Language,
		&user.CreatedAt,
	)

	if err == nil {
		// User exists, update username/display name if changed
		if username != "" || displayName != "" {
			updateQuery := `UPDATE users SET username = ?, display_name = ? WHERE telegram_id = ?`
			_, err := conn.Exec(updateQuery,
				sql.NullString{String: username, Valid: username != ""},
				sql.NullString{String: displayName, Valid: displayName != ""},
				telegramID,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to update user: %w", err)
			}
		}
		return &user, nil
	}

	if err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	// User doesn't exist, create new one
	insertQuery := `INSERT INTO users (telegram_id, username, display_name, language, created_at) 
	                VALUES (?, ?, ?, ?, ?)`

	now := time.Now()
	// Default to English
	defaultLang := i18n.LanguageEnglish
	_, err = conn.Exec(insertQuery,
		telegramID,
		sql.NullString{String: username, Valid: username != ""},
		sql.NullString{String: displayName, Valid: displayName != ""},
		string(defaultLang),
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &database.User{
		TelegramID:  telegramID,
		Username:    sql.NullString{String: username, Valid: username != ""},
		DisplayName: sql.NullString{String: displayName, Valid: displayName != ""},
		Language:    sql.NullString{String: string(defaultLang), Valid: true},
		CreatedAt:   now,
	}, nil
}

// GetUserByTelegramID gets a user by their Telegram ID
func (s *UserService) GetUserByTelegramID(telegramID int64) (*database.User, error) {
	conn := s.db.GetConn()

	var user database.User
	query := `SELECT telegram_id, username, display_name, language, created_at 
	          FROM users WHERE telegram_id = ?`

	err := conn.QueryRow(query, telegramID).Scan(
		&user.TelegramID,
		&user.Username,
		&user.DisplayName,
		&user.Language,
		&user.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	return &user, nil
}

// UpdateUserLanguage updates a user's language preference
func (s *UserService) UpdateUserLanguage(telegramID int64, language i18n.Language) error {
	conn := s.db.GetConn()
	query := `UPDATE users SET language = ? WHERE telegram_id = ?`
	_, err := conn.Exec(query, string(language), telegramID)
	if err != nil {
		return fmt.Errorf("failed to update user language: %w", err)
	}
	return nil
}
