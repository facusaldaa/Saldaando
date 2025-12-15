package service

import (
	"botGastosPareja/internal/database"
	"fmt"
	"time"
)

// SettlementResult represents the result of a settlement calculation
type SettlementResult struct {
	LobbyID               int64
	AccountType           string
	PeriodStart           time.Time
	PeriodEnd             time.Time
	User1ID               int64
	User2ID               int64
	User1TotalSpent       float64
	User2TotalSpent       float64
	TotalExpenses         float64
	User1Expected         float64
	User2Expected         float64
	User1SalaryPercentage float64 // Actual salary percentage from lobby
	User2SalaryPercentage float64 // Actual salary percentage from lobby
	User1Debt             float64 // Positive = user1 owes, Negative = user2 owes user1
	User2Debt             float64 // Positive = user2 owes, Negative = user1 owes user2
	Expenses              []*database.Expense
}

// SettlementService handles settlement calculations
type SettlementService struct {
	db             *database.DB
	expenseService *ExpenseService
	lobbyService   *LobbyService
}

// NewSettlementService creates a new settlement service
func NewSettlementService(db *database.DB, expenseService *ExpenseService, lobbyService *LobbyService) *SettlementService {
	return &SettlementService{
		db:             db,
		expenseService: expenseService,
		lobbyService:   lobbyService,
	}
}

// CalculateSettlement calculates settlement for a lobby within a date range
func (s *SettlementService) CalculateSettlement(lobbyID int64, startDate *time.Time, endDate *time.Time) (*SettlementResult, error) {
	// Get lobby
	lobby, err := s.lobbyService.GetLobbyByID(lobbyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get lobby: %w", err)
	}
	if lobby == nil {
		return nil, fmt.Errorf("lobby not found")
	}

	// Get expenses for the period
	expenses, err := s.expenseService.GetExpensesByLobby(lobbyID, startDate, endDate, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get expenses: %w", err)
	}

	result := &SettlementResult{
		LobbyID:               lobbyID,
		AccountType:           lobby.AccountType,
		User1ID:               lobby.User1TelegramID,
		User2ID:               lobby.User2TelegramID,
		User1SalaryPercentage: lobby.User1SalaryPercentage,
		User2SalaryPercentage: lobby.User2SalaryPercentage,
		Expenses:              expenses,
	}

	if startDate != nil {
		result.PeriodStart = *startDate
	}
	if endDate != nil {
		result.PeriodEnd = *endDate
	}

	// Calculate totals per user
	for _, expense := range expenses {
		result.TotalExpenses += expense.Amount
		if expense.SpenderTelegramID == lobby.User1TelegramID {
			result.User1TotalSpent += expense.Amount
		} else if expense.SpenderTelegramID == lobby.User2TelegramID {
			result.User2TotalSpent += expense.Amount
		}
	}

	// Calculate settlement based on account type
	if lobby.AccountType == "separate" {
		// Equal split
		expectedPerPerson := result.TotalExpenses / 2.0
		result.User1Debt = expectedPerPerson - result.User1TotalSpent
		result.User2Debt = expectedPerPerson - result.User2TotalSpent
		result.User1Expected = expectedPerPerson
		result.User2Expected = expectedPerPerson
	} else {
		// Shared account - based on salary percentage
		result.User1Expected = result.TotalExpenses * lobby.User1SalaryPercentage
		result.User2Expected = result.TotalExpenses * lobby.User2SalaryPercentage
		result.User1Debt = result.User1Expected - result.User1TotalSpent
		result.User2Debt = result.User2Expected - result.User2TotalSpent
	}

	return result, nil
}

// CalculateBillingSettlement calculates settlement for a specific billing period
func (s *SettlementService) CalculateBillingSettlement(lobbyID int64, paymentMethodID int64, periodStart, periodEnd time.Time) (*SettlementResult, error) {
	// Get lobby
	lobby, err := s.lobbyService.GetLobbyByID(lobbyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get lobby: %w", err)
	}
	if lobby == nil {
		return nil, fmt.Errorf("lobby not found")
	}

	// Get expenses for billing period
	expenses, err := s.expenseService.GetExpensesByBillingPeriod(lobbyID, paymentMethodID, periodStart, periodEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to get expenses: %w", err)
	}

	result := &SettlementResult{
		LobbyID:               lobbyID,
		AccountType:           lobby.AccountType,
		User1ID:               lobby.User1TelegramID,
		User2ID:               lobby.User2TelegramID,
		PeriodStart:           periodStart,
		PeriodEnd:             periodEnd,
		User1SalaryPercentage: lobby.User1SalaryPercentage,
		User2SalaryPercentage: lobby.User2SalaryPercentage,
		Expenses:              expenses,
	}

	// Calculate totals per user
	for _, expense := range expenses {
		result.TotalExpenses += expense.Amount
		if expense.SpenderTelegramID == lobby.User1TelegramID {
			result.User1TotalSpent += expense.Amount
		} else if expense.SpenderTelegramID == lobby.User2TelegramID {
			result.User2TotalSpent += expense.Amount
		}
	}

	// Calculate settlement based on account type
	if lobby.AccountType == "separate" {
		expectedPerPerson := result.TotalExpenses / 2.0
		result.User1Debt = expectedPerPerson - result.User1TotalSpent
		result.User2Debt = expectedPerPerson - result.User2TotalSpent
		result.User1Expected = expectedPerPerson
		result.User2Expected = expectedPerPerson
	} else {
		result.User1Expected = result.TotalExpenses * lobby.User1SalaryPercentage
		result.User2Expected = result.TotalExpenses * lobby.User2SalaryPercentage
		result.User1Debt = result.User1Expected - result.User1TotalSpent
		result.User2Debt = result.User2Expected - result.User2TotalSpent
	}

	return result, nil
}
