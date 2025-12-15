package service

import (
	"botGastosPareja/internal/database"
	"botGastosPareja/pkg/utils"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// ExpenseService handles expense operations
type ExpenseService struct {
	db *database.DB
}

// NewExpenseService creates a new expense service
func NewExpenseService(db *database.DB) *ExpenseService {
	return &ExpenseService{db: db}
}

// CreateExpense creates a new expense
func (s *ExpenseService) CreateExpense(lobbyID int64, spenderTelegramID int64, amount float64, description string, category string, expenseDate time.Time, paymentMethodID *int64) (*database.Expense, error) {
	conn := s.db.GetConn()

	var billingPeriodStart, billingPeriodEnd sql.NullTime

	// Calculate billing period if payment method is provided
	if paymentMethodID != nil {
		pmService := NewPaymentMethodService(s.db)
		pm, err := pmService.GetPaymentMethodByID(*paymentMethodID)
		if err != nil {
			return nil, fmt.Errorf("failed to get payment method: %w", err)
		}
		if pm != nil && pm.ClosingDay.Valid {
			start, end := utils.CalculateBillingPeriod(expenseDate, pm.ClosingDay.Int64)
			billingPeriodStart = sql.NullTime{Time: start, Valid: true}
			billingPeriodEnd = sql.NullTime{Time: end, Valid: true}
		}
	}

	var descNull sql.NullString
	if description != "" {
		descNull = sql.NullString{String: description, Valid: true}
	}

	var catNull sql.NullString
	if category != "" {
		catNull = sql.NullString{String: category, Valid: true}
	}

	var pmIDNull sql.NullInt64
	if paymentMethodID != nil {
		pmIDNull = sql.NullInt64{Int64: *paymentMethodID, Valid: true}
	}

	query := `INSERT INTO expenses 
	          (lobby_id, spender_telegram_id, payment_method_id, amount, description, 
	           category, expense_date, billing_period_start, billing_period_end, created_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	result, err := conn.Exec(query,
		lobbyID,
		spenderTelegramID,
		pmIDNull,
		amount,
		descNull,
		catNull,
		expenseDate,
		billingPeriodStart,
		billingPeriodEnd,
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create expense: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get expense ID: %w", err)
	}

	return &database.Expense{
		ID:                id,
		LobbyID:           lobbyID,
		SpenderTelegramID: spenderTelegramID,
		PaymentMethodID:   pmIDNull,
		Amount:            amount,
		Description:       descNull,
		Category:          catNull,
		ExpenseDate:       expenseDate,
		BillingPeriodStart: billingPeriodStart,
		BillingPeriodEnd:   billingPeriodEnd,
		CreatedAt:         now,
	}, nil
}

// GetExpensesByLobby gets expenses for a lobby with optional filters
func (s *ExpenseService) GetExpensesByLobby(lobbyID int64, startDate *time.Time, endDate *time.Time, paymentMethodID *int64) ([]*database.Expense, error) {
	conn := s.db.GetConn()

	query := `SELECT id, lobby_id, spender_telegram_id, payment_method_id, amount, 
	          description, category, expense_date, billing_period_start, 
	          billing_period_end, created_at
	          FROM expenses WHERE lobby_id = ?`

	args := []interface{}{lobbyID}

	if startDate != nil {
		query += " AND expense_date >= ?"
		args = append(args, *startDate)
	}

	if endDate != nil {
		query += " AND expense_date <= ?"
		args = append(args, *endDate)
	}

	if paymentMethodID != nil {
		query += " AND payment_method_id = ?"
		args = append(args, *paymentMethodID)
	}

	query += " ORDER BY expense_date DESC, created_at DESC"

	rows, err := conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query expenses: %w", err)
	}
	defer rows.Close()

	var expenses []*database.Expense
	for rows.Next() {
		var expense database.Expense
		err := rows.Scan(
			&expense.ID,
			&expense.LobbyID,
			&expense.SpenderTelegramID,
			&expense.PaymentMethodID,
			&expense.Amount,
			&expense.Description,
			&expense.Category,
			&expense.ExpenseDate,
			&expense.BillingPeriodStart,
			&expense.BillingPeriodEnd,
			&expense.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan expense: %w", err)
		}
		expenses = append(expenses, &expense)
	}

	return expenses, nil
}

// GetExpensesByBillingPeriod gets expenses for a specific billing period
func (s *ExpenseService) GetExpensesByBillingPeriod(lobbyID int64, paymentMethodID int64, periodStart, periodEnd time.Time) ([]*database.Expense, error) {
	conn := s.db.GetConn()

	query := `SELECT id, lobby_id, spender_telegram_id, payment_method_id, amount, 
	          description, category, expense_date, billing_period_start, 
	          billing_period_end, created_at
	          FROM expenses 
	          WHERE lobby_id = ? AND payment_method_id = ?
	          AND billing_period_start >= ? AND billing_period_end <= ?
	          ORDER BY expense_date DESC`

	rows, err := conn.Query(query, lobbyID, paymentMethodID, periodStart, periodEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to query expenses: %w", err)
	}
	defer rows.Close()

	var expenses []*database.Expense
	for rows.Next() {
		var expense database.Expense
		err := rows.Scan(
			&expense.ID,
			&expense.LobbyID,
			&expense.SpenderTelegramID,
			&expense.PaymentMethodID,
			&expense.Amount,
			&expense.Description,
			&expense.Category,
			&expense.ExpenseDate,
			&expense.BillingPeriodStart,
			&expense.BillingPeriodEnd,
			&expense.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan expense: %w", err)
		}
		expenses = append(expenses, &expense)
	}

	return expenses, nil
}

// GetExpenseByID gets an expense by ID
func (s *ExpenseService) GetExpenseByID(id int64) (*database.Expense, error) {
	conn := s.db.GetConn()

	var expense database.Expense
	query := `SELECT id, lobby_id, spender_telegram_id, payment_method_id, amount, 
	          description, category, expense_date, billing_period_start, 
	          billing_period_end, created_at
	          FROM expenses WHERE id = ?`

	err := conn.QueryRow(query, id).Scan(
		&expense.ID,
		&expense.LobbyID,
		&expense.SpenderTelegramID,
		&expense.PaymentMethodID,
		&expense.Amount,
		&expense.Description,
		&expense.Category,
		&expense.ExpenseDate,
		&expense.BillingPeriodStart,
		&expense.BillingPeriodEnd,
		&expense.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query expense: %w", err)
	}

	return &expense, nil
}

// UpdateExpense updates an expense
func (s *ExpenseService) UpdateExpense(id int64, amount *float64, description *string, category *string, expenseDate *time.Time, paymentMethodID *int64) error {
	conn := s.db.GetConn()

	updates := []string{}
	args := []interface{}{}

	if amount != nil {
		updates = append(updates, "amount = ?")
		args = append(args, *amount)
	}

	if description != nil {
		updates = append(updates, "description = ?")
		args = append(args, sql.NullString{String: *description, Valid: *description != ""})
	}

	if category != nil {
		updates = append(updates, "category = ?")
		args = append(args, sql.NullString{String: *category, Valid: *category != ""})
	}

	if expenseDate != nil {
		updates = append(updates, "expense_date = ?")
		args = append(args, *expenseDate)
	}

	if paymentMethodID != nil {
		updates = append(updates, "payment_method_id = ?")
		args = append(args, sql.NullInt64{Int64: *paymentMethodID, Valid: true})
		
		// Recalculate billing period
		pmService := NewPaymentMethodService(s.db)
		pm, err := pmService.GetPaymentMethodByID(*paymentMethodID)
		if err == nil && pm != nil && pm.ClosingDay.Valid {
			date := expenseDate
			if date == nil {
				// Get current expense date
				expense, err := s.GetExpenseByID(id)
				if err == nil && expense != nil {
					date = &expense.ExpenseDate
				}
			}
			if date != nil {
				start, end := utils.CalculateBillingPeriod(*date, pm.ClosingDay.Int64)
				updates = append(updates, "billing_period_start = ?")
				updates = append(updates, "billing_period_end = ?")
				args = append(args, start, end)
			}
		}
	}

	if len(updates) == 0 {
		return nil // Nothing to update
	}

	args = append(args, id)
	query := fmt.Sprintf("UPDATE expenses SET %s WHERE id = ?", 
		strings.Join(updates, ", "))

	_, err := conn.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update expense: %w", err)
	}

	return nil
}

// DeleteExpense deletes an expense
func (s *ExpenseService) DeleteExpense(id int64) error {
	conn := s.db.GetConn()
	query := `DELETE FROM expenses WHERE id = ?`
	_, err := conn.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete expense: %w", err)
	}
	return nil
}

