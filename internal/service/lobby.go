package service

import (
	"botGastosPareja/internal/database"
	"botGastosPareja/pkg/utils"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"
)

// LobbyService handles lobby-related operations
type LobbyService struct {
	db *database.DB
}

// NewLobbyService creates a new lobby service
func NewLobbyService(db *database.DB) *LobbyService {
	return &LobbyService{db: db}
}

// GetLobbyByID gets a lobby by ID
func (s *LobbyService) GetLobbyByID(lobbyID int64) (*database.Lobby, error) {
	conn := s.db.GetConn()

	// Check if new columns exist
	var hasInviteToken, hasGroupChatID bool
	_ = conn.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('lobbies') WHERE name='invite_token'`).Scan(&hasInviteToken)
	_ = conn.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('lobbies') WHERE name='group_chat_id'`).Scan(&hasGroupChatID)

	var lobby database.Lobby
	var err error

	var user2ID sql.NullInt64

	if hasInviteToken && hasGroupChatID {
		// New schema with all columns
		query := `SELECT id, user1_telegram_id, user2_telegram_id, account_type, 
		          user1_salary_percentage, user2_salary_percentage, invite_token, 
		          group_chat_id, created_at
		          FROM lobbies WHERE id = ?`
		err = conn.QueryRow(query, lobbyID).Scan(
			&lobby.ID,
			&lobby.User1TelegramID,
			&user2ID,
			&lobby.AccountType,
			&lobby.User1SalaryPercentage,
			&lobby.User2SalaryPercentage,
			&lobby.InviteToken,
			&lobby.GroupChatID,
			&lobby.CreatedAt,
		)
	} else {
		// Old schema without new columns
		query := `SELECT id, user1_telegram_id, user2_telegram_id, account_type, 
		          user1_salary_percentage, user2_salary_percentage, created_at
		          FROM lobbies WHERE id = ?`
		err = conn.QueryRow(query, lobbyID).Scan(
			&lobby.ID,
			&lobby.User1TelegramID,
			&user2ID,
			&lobby.AccountType,
			&lobby.User1SalaryPercentage,
			&lobby.User2SalaryPercentage,
			&lobby.CreatedAt,
		)
		// Set defaults for missing columns
		lobby.InviteToken = sql.NullString{Valid: false}
		lobby.GroupChatID = sql.NullInt64{Valid: false}
	}

	// Convert NullInt64 to int64 (0 if NULL)
	if user2ID.Valid {
		lobby.User2TelegramID = user2ID.Int64
	} else {
		lobby.User2TelegramID = 0
	}

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query lobby: %w", err)
	}

	return &lobby, nil
}

// GetLobbyByUserID gets a lobby for a user (if they're in one)
func (s *LobbyService) GetLobbyByUserID(userID int64) (*database.Lobby, error) {
	conn := s.db.GetConn()

	// Check if new columns exist
	var hasInviteToken, hasGroupChatID bool
	_ = conn.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('lobbies') WHERE name='invite_token'`).Scan(&hasInviteToken)
	_ = conn.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('lobbies') WHERE name='group_chat_id'`).Scan(&hasGroupChatID)

	var lobby database.Lobby
	var query string

	var user2ID sql.NullInt64

	if hasInviteToken && hasGroupChatID {
		// New schema with all columns
		query = `SELECT id, user1_telegram_id, user2_telegram_id, account_type, 
		          user1_salary_percentage, user2_salary_percentage, invite_token,
		          group_chat_id, created_at
		          FROM lobbies 
		          WHERE user1_telegram_id = ? OR user2_telegram_id = ?`
		err := conn.QueryRow(query, userID, userID).Scan(
			&lobby.ID,
			&lobby.User1TelegramID,
			&user2ID,
			&lobby.AccountType,
			&lobby.User1SalaryPercentage,
			&lobby.User2SalaryPercentage,
			&lobby.InviteToken,
			&lobby.GroupChatID,
			&lobby.CreatedAt,
		)
		if err == sql.ErrNoRows {
			return nil, nil
		}
		if err != nil {
			return nil, fmt.Errorf("failed to query lobby: %w", err)
		}
	} else {
		// Old schema without new columns
		query = `SELECT id, user1_telegram_id, user2_telegram_id, account_type, 
		          user1_salary_percentage, user2_salary_percentage, created_at
		          FROM lobbies 
		          WHERE user1_telegram_id = ? OR user2_telegram_id = ?`
		err := conn.QueryRow(query, userID, userID).Scan(
			&lobby.ID,
			&lobby.User1TelegramID,
			&user2ID,
			&lobby.AccountType,
			&lobby.User1SalaryPercentage,
			&lobby.User2SalaryPercentage,
			&lobby.CreatedAt,
		)
		if err == sql.ErrNoRows {
			return nil, nil
		}
		if err != nil {
			return nil, fmt.Errorf("failed to query lobby: %w", err)
		}
		// Set defaults for missing columns
		lobby.InviteToken = sql.NullString{Valid: false}
		lobby.GroupChatID = sql.NullInt64{Valid: false}
	}

	// Convert NullInt64 to int64 (0 if NULL)
	if user2ID.Valid {
		lobby.User2TelegramID = user2ID.Int64
	} else {
		lobby.User2TelegramID = 0
	}

	return &lobby, nil
}

// GetLobbyByUserIDAndGroup gets a lobby for a user in a specific group (or private if groupChatID is nil)
func (s *LobbyService) GetLobbyByUserIDAndGroup(userID int64, groupChatID *int64) (*database.Lobby, error) {
	conn := s.db.GetConn()

	// Check if new columns exist
	var hasGroupChatID bool
	_ = conn.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('lobbies') WHERE name='group_chat_id'`).Scan(&hasGroupChatID)

	if !hasGroupChatID {
		// Old schema - fallback to regular lookup
		return s.GetLobbyByUserID(userID)
	}

	var lobby database.Lobby
	var query string
	var err error

	if groupChatID != nil {
		// Look for lobby in this specific group
		query = `SELECT id, user1_telegram_id, user2_telegram_id, account_type, 
		          user1_salary_percentage, user2_salary_percentage, invite_token,
		          group_chat_id, created_at
		          FROM lobbies 
		          WHERE (user1_telegram_id = ? OR user2_telegram_id = ?)
		          AND group_chat_id = ?`

		// Log for debugging
		log.Printf("DEBUG GetLobbyByUserIDAndGroup: userID=%d, groupChatID=%d", userID, *groupChatID)

		// Scan user2_telegram_id as NullInt64 since it can be NULL
		var user2ID sql.NullInt64
		err = conn.QueryRow(query, userID, userID, *groupChatID).Scan(
			&lobby.ID,
			&lobby.User1TelegramID,
			&user2ID,
			&lobby.AccountType,
			&lobby.User1SalaryPercentage,
			&lobby.User2SalaryPercentage,
			&lobby.InviteToken,
			&lobby.GroupChatID,
			&lobby.CreatedAt,
		)

		// Convert NullInt64 to int64 (0 if NULL)
		if user2ID.Valid {
			lobby.User2TelegramID = user2ID.Int64
		} else {
			lobby.User2TelegramID = 0
		}

		if err == sql.ErrNoRows {
			log.Printf("DEBUG: No lobby found for userID=%d, groupChatID=%d", userID, *groupChatID)
			// Let's also check what lobbies exist for this user
			checkQuery := `SELECT id, user1_telegram_id, user2_telegram_id, group_chat_id FROM lobbies WHERE user1_telegram_id = ? OR user2_telegram_id = ?`
			rows, _ := conn.Query(checkQuery, userID, userID)
			if rows != nil {
				defer rows.Close()
				log.Printf("DEBUG: Checking all lobbies for userID=%d:", userID)
				for rows.Next() {
					var lid, u1, u2 int64
					var gcid sql.NullInt64
					rows.Scan(&lid, &u1, &u2, &gcid)
					log.Printf("  Lobby ID=%d, User1=%d, User2=%d, GroupChatID=%v (Valid=%v)", lid, u1, u2, gcid.Int64, gcid.Valid)
				}
			}
		}
	} else {
		// Look for private lobby (no group_chat_id)
		query = `SELECT id, user1_telegram_id, user2_telegram_id, account_type, 
		          user1_salary_percentage, user2_salary_percentage, invite_token,
		          group_chat_id, created_at
		          FROM lobbies 
		          WHERE (user1_telegram_id = ? OR user2_telegram_id = ?)
		          AND (group_chat_id IS NULL)`
		// Scan user2_telegram_id as NullInt64 since it can be NULL
		var user2ID sql.NullInt64
		err = conn.QueryRow(query, userID, userID).Scan(
			&lobby.ID,
			&lobby.User1TelegramID,
			&user2ID,
			&lobby.AccountType,
			&lobby.User1SalaryPercentage,
			&lobby.User2SalaryPercentage,
			&lobby.InviteToken,
			&lobby.GroupChatID,
			&lobby.CreatedAt,
		)

		// Convert NullInt64 to int64 (0 if NULL)
		if user2ID.Valid {
			lobby.User2TelegramID = user2ID.Int64
		} else {
			lobby.User2TelegramID = 0
		}
	}

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		log.Printf("DEBUG GetLobbyByUserIDAndGroup ERROR: %v", err)
		return nil, fmt.Errorf("failed to query lobby: %w", err)
	}

	log.Printf("DEBUG: Found lobby ID=%d for userID=%d, groupChatID=%v", lobby.ID, userID, groupChatID)
	return &lobby, nil
}

// GetLobbyByGroupChatID gets a lobby for a specific group/channel
func (s *LobbyService) GetLobbyByGroupChatID(groupChatID int64) (*database.Lobby, error) {
	conn := s.db.GetConn()

	// Check if new columns exist
	var hasGroupChatID bool
	_ = conn.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('lobbies') WHERE name='group_chat_id'`).Scan(&hasGroupChatID)

	if !hasGroupChatID {
		return nil, nil
	}

	var lobby database.Lobby
	var user2ID sql.NullInt64
	query := `SELECT id, user1_telegram_id, user2_telegram_id, account_type, 
	          user1_salary_percentage, user2_salary_percentage, invite_token,
	          group_chat_id, created_at
	          FROM lobbies 
	          WHERE group_chat_id = ?
	          ORDER BY created_at ASC
	          LIMIT 1`

	err := conn.QueryRow(query, groupChatID).Scan(
		&lobby.ID,
		&lobby.User1TelegramID,
		&user2ID,
		&lobby.AccountType,
		&lobby.User1SalaryPercentage,
		&lobby.User2SalaryPercentage,
		&lobby.InviteToken,
		&lobby.GroupChatID,
		&lobby.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query lobby: %w", err)
	}

	// Convert NullInt64 to int64 (0 if NULL)
	if user2ID.Valid {
		lobby.User2TelegramID = user2ID.Int64
	} else {
		lobby.User2TelegramID = 0
	}

	return &lobby, nil
}

// CreateLobby creates a new lobby with one user
func (s *LobbyService) CreateLobby(userID int64, accountType string, groupChatID *int64) (*database.Lobby, error) {
	conn := s.db.GetConn()

	// Validate account type
	if accountType != "separate" && accountType != "shared" {
		accountType = "separate" // Default
	}

	// Generate secure invitation token
	inviteToken, err := utils.GenerateInviteToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate invite token: %w", err)
	}

	var groupChatIDNull sql.NullInt64
	if groupChatID != nil {
		groupChatIDNull = sql.NullInt64{Int64: *groupChatID, Valid: true}
		log.Printf("DEBUG CreateLobby: Creating lobby for userID=%d, groupChatID=%d", userID, *groupChatID)
	} else {
		log.Printf("DEBUG CreateLobby: Creating private lobby for userID=%d", userID)
	}

	query := `INSERT INTO lobbies (user1_telegram_id, user2_telegram_id, account_type, 
	          user1_salary_percentage, user2_salary_percentage, invite_token, 
	          group_chat_id, created_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	result, err := conn.Exec(query,
		userID,
		nil, // user2 will be set when they join
		accountType,
		0.5, // Default equal split
		0.5,
		inviteToken,
		groupChatIDNull,
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create lobby: %w", err)
	}

	lobbyID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get lobby ID: %w", err)
	}

	log.Printf("DEBUG CreateLobby: Created lobby ID=%d for userID=%d, groupChatID=%v", lobbyID, userID, groupChatIDNull)

	return &database.Lobby{
		ID:                    lobbyID,
		User1TelegramID:       userID,
		User2TelegramID:       0, // Not set yet
		AccountType:           accountType,
		User1SalaryPercentage: 0.5,
		User2SalaryPercentage: 0.5,
		InviteToken:           sql.NullString{String: inviteToken, Valid: true},
		GroupChatID:           groupChatIDNull,
		CreatedAt:             now,
	}, nil
}

// GetLobbyByInviteToken gets a lobby by invitation token
func (s *LobbyService) GetLobbyByInviteToken(token string) (*database.Lobby, error) {
	conn := s.db.GetConn()

	// Remove formatting if present
	cleanToken := utils.ParseInviteToken(token)

	var lobby database.Lobby
	var user2ID sql.NullInt64
	query := `SELECT id, user1_telegram_id, user2_telegram_id, account_type, 
	          user1_salary_percentage, user2_salary_percentage, invite_token,
	          group_chat_id, created_at
	          FROM lobbies WHERE invite_token = ?`

	err := conn.QueryRow(query, cleanToken).Scan(
		&lobby.ID,
		&lobby.User1TelegramID,
		&user2ID,
		&lobby.AccountType,
		&lobby.User1SalaryPercentage,
		&lobby.User2SalaryPercentage,
		&lobby.InviteToken,
		&lobby.GroupChatID,
		&lobby.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query lobby: %w", err)
	}

	// Convert NullInt64 to int64 (0 if NULL)
	if user2ID.Valid {
		lobby.User2TelegramID = user2ID.Int64
	} else {
		lobby.User2TelegramID = 0
	}

	return &lobby, nil
}

// RegenerateInviteToken generates a new invitation token for a lobby
func (s *LobbyService) RegenerateInviteToken(lobbyID int64) (string, error) {
	conn := s.db.GetConn()

	newToken, err := utils.GenerateInviteToken()
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	query := `UPDATE lobbies SET invite_token = ? WHERE id = ?`
	_, err = conn.Exec(query, newToken, lobbyID)
	if err != nil {
		return "", fmt.Errorf("failed to update invite token: %w", err)
	}

	return newToken, nil
}

// JoinLobbyByToken allows a second user to join an existing lobby by token
// groupChatID is used to validate that the lobby is for the same group (or private)
func (s *LobbyService) JoinLobbyByToken(inviteToken string, userID int64, groupChatID *int64) error {
	conn := s.db.GetConn()

	// Get lobby by token
	lobby, err := s.GetLobbyByInviteToken(inviteToken)
	if err != nil {
		return fmt.Errorf("failed to get lobby: %w", err)
	}
	if lobby == nil {
		return fmt.Errorf("invalid invitation token")
	}

	// Validate group chat ID matches
	if groupChatID != nil {
		// Joining in a group - lobby must be for this group
		if !lobby.GroupChatID.Valid || lobby.GroupChatID.Int64 != *groupChatID {
			return fmt.Errorf("this invitation token is for a different chat")
		}
	} else {
		// Joining in private - lobby must be private (no group)
		if lobby.GroupChatID.Valid {
			return fmt.Errorf("this invitation token is for a group chat. Please join from that group")
		}
	}

	// Check if lobby is full
	if lobby.User2TelegramID != 0 {
		return fmt.Errorf("lobby is already full")
	}

	// Check if user is already user1
	if lobby.User1TelegramID == userID {
		return fmt.Errorf("you are already in this lobby")
	}

	// Update lobby with user2
	updateQuery := `UPDATE lobbies SET user2_telegram_id = ? WHERE id = ?`
	_, err = conn.Exec(updateQuery, userID, lobby.ID)
	if err != nil {
		return fmt.Errorf("failed to join lobby: %w", err)
	}

	return nil
}

// JoinLobbyDirectly allows a second user to join an existing lobby directly (without token)
// Used when a user joins a group/channel that already has a lobby
func (s *LobbyService) JoinLobbyDirectly(lobbyID int64, userID int64) error {
	conn := s.db.GetConn()

	// Get the lobby
	lobby, err := s.GetLobbyByID(lobbyID)
	if err != nil {
		return fmt.Errorf("failed to get lobby: %w", err)
	}
	if lobby == nil {
		return fmt.Errorf("lobby not found")
	}

	// Check if lobby is full
	if lobby.User2TelegramID != 0 {
		return fmt.Errorf("lobby is already full")
	}

	// Check if user is already user1
	if lobby.User1TelegramID == userID {
		return fmt.Errorf("you are already in this lobby")
	}

	// Update lobby with user2
	updateQuery := `UPDATE lobbies SET user2_telegram_id = ? WHERE id = ?`
	_, err = conn.Exec(updateQuery, userID, lobby.ID)
	if err != nil {
		return fmt.Errorf("failed to join lobby: %w", err)
	}

	return nil
}

// JoinLobby allows a second user to join an existing lobby (deprecated - use JoinLobbyByToken)
func (s *LobbyService) JoinLobby(lobbyID int64, userID int64) error {
	conn := s.db.GetConn()

	// Check if lobby exists and has space
	var currentUser2 sql.NullInt64
	query := `SELECT user2_telegram_id FROM lobbies WHERE id = ?`
	err := conn.QueryRow(query, lobbyID).Scan(&currentUser2)
	if err == sql.ErrNoRows {
		return fmt.Errorf("lobby not found")
	}
	if err != nil {
		return fmt.Errorf("failed to query lobby: %w", err)
	}

	// Check if lobby is full
	if currentUser2.Valid {
		return fmt.Errorf("lobby is already full")
	}

	// Check if user is already user1
	var user1ID int64
	err = conn.QueryRow(`SELECT user1_telegram_id FROM lobbies WHERE id = ?`, lobbyID).Scan(&user1ID)
	if err != nil {
		return fmt.Errorf("failed to query lobby: %w", err)
	}
	if user1ID == userID {
		return fmt.Errorf("you are already in this lobby")
	}

	// Update lobby with user2
	updateQuery := `UPDATE lobbies SET user2_telegram_id = ? WHERE id = ?`
	_, err = conn.Exec(updateQuery, userID, lobbyID)
	if err != nil {
		return fmt.Errorf("failed to join lobby: %w", err)
	}

	return nil
}

// UpdateLobbySettings updates lobby configuration
func (s *LobbyService) UpdateLobbySettings(lobbyID int64, accountType *string, user1SalaryPct *float64, user2SalaryPct *float64) error {
	conn := s.db.GetConn()

	updates := []string{}
	args := []interface{}{}

	if accountType != nil {
		if *accountType != "separate" && *accountType != "shared" {
			return fmt.Errorf("invalid account type: %s", *accountType)
		}
		updates = append(updates, "account_type = ?")
		args = append(args, *accountType)
	}

	if user1SalaryPct != nil {
		if *user1SalaryPct < 0 || *user1SalaryPct > 1 {
			return fmt.Errorf("salary percentage must be between 0 and 1")
		}
		updates = append(updates, "user1_salary_percentage = ?")
		args = append(args, *user1SalaryPct)
	}

	if user2SalaryPct != nil {
		if *user2SalaryPct < 0 || *user2SalaryPct > 1 {
			return fmt.Errorf("salary percentage must be between 0 and 1")
		}
		updates = append(updates, "user2_salary_percentage = ?")
		args = append(args, *user2SalaryPct)
	}

	if len(updates) == 0 {
		return nil // Nothing to update
	}

	args = append(args, lobbyID)
	query := fmt.Sprintf("UPDATE lobbies SET %s WHERE id = ?",
		strings.Join(updates, ", "))

	_, err := conn.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update lobby settings: %w", err)
	}

	return nil
}
