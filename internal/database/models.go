package database

import (
	"database/sql"
	"time"
)

// User represents a Telegram user
type User struct {
	TelegramID  int64
	Username    sql.NullString
	DisplayName sql.NullString
	Language    sql.NullString // Language preference (e.g., "en", "es_AR")
	CreatedAt   time.Time
}

// Lobby represents a couple's shared expense tracking space
type Lobby struct {
	ID                    int64
	User1TelegramID       int64
	User2TelegramID       int64
	AccountType           string // "separate" or "shared"
	User1SalaryPercentage float64
	User2SalaryPercentage float64
	InviteToken           sql.NullString // Secure invitation token
	GroupChatID           sql.NullInt64  // Telegram group/channel ID (optional)
	CreatedAt             time.Time
}

// Category represents an expense category
type Category struct {
	ID        int64
	LobbyID   int64
	Name      string
	IsDefault bool
}

// PaymentMethod represents a payment method (credit card, cash, etc.)
type PaymentMethod struct {
	ID               int64
	LobbyID          int64
	Name             string
	Type             string        // "credit_card", "debit_card", "cash", "bank_transfer", "other"
	OwnerTelegramID  sql.NullInt64 // NULL if shared
	ClosingDay       sql.NullInt64 // Day of month when statement closes (1-31)
	BillingCycleDays int64
	IsActive         bool
	CreatedAt        time.Time
}

// Expense represents a single expense entry
type Expense struct {
	ID                 int64
	LobbyID            int64
	SpenderTelegramID  int64
	PaymentMethodID    sql.NullInt64
	Amount             float64
	Description        sql.NullString
	Category           sql.NullString
	ExpenseDate        time.Time
	BillingPeriodStart sql.NullTime
	BillingPeriodEnd   sql.NullTime
	CreatedAt          time.Time
}
