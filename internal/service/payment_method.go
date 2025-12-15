package service

import (
	"botGastosPareja/internal/database"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// PaymentMethodService handles payment method operations
type PaymentMethodService struct {
	db *database.DB
}

// NewPaymentMethodService creates a new payment method service
func NewPaymentMethodService(db *database.DB) *PaymentMethodService {
	return &PaymentMethodService{db: db}
}

// normalizePaymentMethodType normalizes payment method type (handles Spanish aliases)
func normalizePaymentMethodType(methodType string) string {
	methodType = strings.ToLower(methodType)

	// Map Spanish aliases to English types
	aliases := map[string]string{
		"tarjetacredito":         "credit_card",
		"tarjeta_credito":        "credit_card",
		"tarjetadebito":          "debit_card",
		"tarjeta_debito":         "debit_card",
		"efectivo":               "cash",
		"transferencia":          "bank_transfer",
		"transferencia_bancaria": "bank_transfer",
		"otro":                   "other",
	}

	if normalized, ok := aliases[methodType]; ok {
		return normalized
	}
	return methodType
}

// CreatePaymentMethod creates a new payment method
func (s *PaymentMethodService) CreatePaymentMethod(lobbyID int64, name string, methodType string, ownerTelegramID *int64, closingDay *int64) (*database.PaymentMethod, error) {
	conn := s.db.GetConn()

	// Normalize method type (handles Spanish aliases)
	methodType = normalizePaymentMethodType(methodType)

	// Validate method type
	validTypes := map[string]bool{
		"credit_card":   true,
		"debit_card":    true,
		"cash":          true,
		"bank_transfer": true,
		"other":         true,
	}
	if !validTypes[methodType] {
		return nil, fmt.Errorf("invalid payment method type: %s", methodType)
	}

	// Validate closing day if provided
	if closingDay != nil {
		if *closingDay < 1 || *closingDay > 31 {
			return nil, fmt.Errorf("closing day must be between 1 and 31")
		}
	}

	var ownerID sql.NullInt64
	if ownerTelegramID != nil {
		ownerID = sql.NullInt64{Int64: *ownerTelegramID, Valid: true}
	}

	var closingDayNull sql.NullInt64
	if closingDay != nil {
		closingDayNull = sql.NullInt64{Int64: *closingDay, Valid: true}
	}

	query := `INSERT INTO payment_methods 
	          (lobby_id, name, type, owner_telegram_id, closing_day, billing_cycle_days, is_active, created_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	result, err := conn.Exec(query,
		lobbyID,
		name,
		methodType,
		ownerID,
		closingDayNull,
		30, // Default billing cycle
		true,
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment method: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get payment method ID: %w", err)
	}

	return &database.PaymentMethod{
		ID:               id,
		LobbyID:          lobbyID,
		Name:             name,
		Type:             methodType,
		OwnerTelegramID:  ownerID,
		ClosingDay:       closingDayNull,
		BillingCycleDays: 30,
		IsActive:         true,
		CreatedAt:        now,
	}, nil
}

// GetPaymentMethodsByLobby gets all payment methods for a lobby
func (s *PaymentMethodService) GetPaymentMethodsByLobby(lobbyID int64, activeOnly bool) ([]*database.PaymentMethod, error) {
	conn := s.db.GetConn()

	query := `SELECT id, lobby_id, name, type, owner_telegram_id, closing_day, 
	          billing_cycle_days, is_active, created_at
	          FROM payment_methods WHERE lobby_id = ?`

	if activeOnly {
		query += " AND is_active = 1"
	}
	query += " ORDER BY name"

	rows, err := conn.Query(query, lobbyID)
	if err != nil {
		return nil, fmt.Errorf("failed to query payment methods: %w", err)
	}
	defer rows.Close()

	var methods []*database.PaymentMethod
	for rows.Next() {
		var method database.PaymentMethod
		err := rows.Scan(
			&method.ID,
			&method.LobbyID,
			&method.Name,
			&method.Type,
			&method.OwnerTelegramID,
			&method.ClosingDay,
			&method.BillingCycleDays,
			&method.IsActive,
			&method.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan payment method: %w", err)
		}
		methods = append(methods, &method)
	}

	return methods, nil
}

// GetPaymentMethodByID gets a payment method by ID
func (s *PaymentMethodService) GetPaymentMethodByID(id int64) (*database.PaymentMethod, error) {
	conn := s.db.GetConn()

	var method database.PaymentMethod
	query := `SELECT id, lobby_id, name, type, owner_telegram_id, closing_day,
	          billing_cycle_days, is_active, created_at
	          FROM payment_methods WHERE id = ?`

	err := conn.QueryRow(query, id).Scan(
		&method.ID,
		&method.LobbyID,
		&method.Name,
		&method.Type,
		&method.OwnerTelegramID,
		&method.ClosingDay,
		&method.BillingCycleDays,
		&method.IsActive,
		&method.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query payment method: %w", err)
	}

	return &method, nil
}

// UpdatePaymentMethod updates a payment method
func (s *PaymentMethodService) UpdatePaymentMethod(id int64, name *string, methodType *string, ownerTelegramID *int64, closingDay *int64, isActive *bool) error {
	conn := s.db.GetConn()

	updates := []string{}
	args := []interface{}{}

	if name != nil {
		updates = append(updates, "name = ?")
		args = append(args, *name)
	}

	if methodType != nil {
		// Normalize method type (handles Spanish aliases)
		normalizedType := normalizePaymentMethodType(*methodType)
		validTypes := map[string]bool{
			"credit_card":   true,
			"debit_card":    true,
			"cash":          true,
			"bank_transfer": true,
			"other":         true,
		}
		if !validTypes[normalizedType] {
			return fmt.Errorf("invalid payment method type: %s", *methodType)
		}
		updates = append(updates, "type = ?")
		args = append(args, normalizedType)
	}

	if ownerTelegramID != nil {
		updates = append(updates, "owner_telegram_id = ?")
		args = append(args, sql.NullInt64{Int64: *ownerTelegramID, Valid: true})
	}

	if closingDay != nil {
		if *closingDay < 1 || *closingDay > 31 {
			return fmt.Errorf("closing day must be between 1 and 31")
		}
		updates = append(updates, "closing_day = ?")
		args = append(args, sql.NullInt64{Int64: *closingDay, Valid: true})
	}

	if isActive != nil {
		updates = append(updates, "is_active = ?")
		args = append(args, *isActive)
	}

	if len(updates) == 0 {
		return nil // Nothing to update
	}

	args = append(args, id)
	query := fmt.Sprintf("UPDATE payment_methods SET %s WHERE id = ?",
		strings.Join(updates, ", "))

	_, err := conn.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update payment method: %w", err)
	}

	return nil
}

// DeletePaymentMethod deletes a payment method (soft delete by setting is_active = false)
func (s *PaymentMethodService) DeletePaymentMethod(id int64) error {
	return s.UpdatePaymentMethod(id, nil, nil, nil, nil, boolPtr(false))
}

func boolPtr(b bool) *bool {
	return &b
}
