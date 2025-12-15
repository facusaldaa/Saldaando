package service

import (
	"botGastosPareja/internal/database"
	"botGastosPareja/pkg/utils"
	"fmt"
	"sort"
	"time"
)

// AnalysisResult represents analysis results
type AnalysisResult struct {
	LobbyID              int64
	CurrentPeriod        time.Time
	PreviousPeriod       time.Time
	CurrentTotal         float64
	PreviousTotal        float64
	ChangePercent        float64
	CategoryChanges      map[string]CategoryChange
	SpendingSpikes       []SpendingSpike
	NewCategories        []string
	DiscontinuedCategories []string
}

// CategoryChange represents changes in category spending
type CategoryChange struct {
	Name         string
	CurrentTotal float64
	PreviousTotal float64
	ChangePercent float64
}

// SpendingSpike represents a detected spending spike
type SpendingSpike struct {
	Category    string
	Amount      float64
	ChangePercent float64
	Period      time.Time
}

// AnalysisService handles spending analysis
type AnalysisService struct {
	db            *database.DB
	expenseService *ExpenseService
}

// NewAnalysisService creates a new analysis service
func NewAnalysisService(db *database.DB, expenseService *ExpenseService) *AnalysisService {
	return &AnalysisService{
		db:            db,
		expenseService: expenseService,
	}
}

// AnalyzeMonthly compares current month with previous month
func (s *AnalysisService) AnalyzeMonthly(lobbyID int64) (*AnalysisResult, error) {
	now := time.Now()
	currentStart, currentEnd := utils.GetMonthStartEnd(now.Year(), now.Month())
	
	prevMonth := now.AddDate(0, -1, 0)
	prevStart, prevEnd := utils.GetMonthStartEnd(prevMonth.Year(), prevMonth.Month())

	currentExpenses, err := s.expenseService.GetExpensesByLobby(lobbyID, &currentStart, &currentEnd, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get current expenses: %w", err)
	}

	previousExpenses, err := s.expenseService.GetExpensesByLobby(lobbyID, &prevStart, &prevEnd, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get previous expenses: %w", err)
	}

	result := &AnalysisResult{
		LobbyID:        lobbyID,
		CurrentPeriod:  currentStart,
		PreviousPeriod: prevStart,
		CategoryChanges: make(map[string]CategoryChange),
		SpendingSpikes: []SpendingSpike{},
		NewCategories:  []string{},
		DiscontinuedCategories: []string{},
	}

	// Calculate totals
	for _, exp := range currentExpenses {
		result.CurrentTotal += exp.Amount
	}
	for _, exp := range previousExpenses {
		result.PreviousTotal += exp.Amount
	}

	// Calculate change percentage
	if result.PreviousTotal > 0 {
		result.ChangePercent = ((result.CurrentTotal - result.PreviousTotal) / result.PreviousTotal) * 100
	} else if result.CurrentTotal > 0 {
		result.ChangePercent = 100 // New spending
	}

	// Analyze by category
	currentCategories := make(map[string]float64)
	previousCategories := make(map[string]float64)

	for _, exp := range currentExpenses {
		if exp.Category.Valid {
			currentCategories[exp.Category.String] += exp.Amount
		}
	}

	for _, exp := range previousExpenses {
		if exp.Category.Valid {
			previousCategories[exp.Category.String] += exp.Amount
		}
	}

	// Find category changes
	allCategories := make(map[string]bool)
	for cat := range currentCategories {
		allCategories[cat] = true
	}
	for cat := range previousCategories {
		allCategories[cat] = true
	}

	for cat := range allCategories {
		currentTotal := currentCategories[cat]
		previousTotal := previousCategories[cat]
		
		if currentTotal > 0 && previousTotal == 0 {
			result.NewCategories = append(result.NewCategories, cat)
		} else if currentTotal == 0 && previousTotal > 0 {
			result.DiscontinuedCategories = append(result.DiscontinuedCategories, cat)
		}

		var changePercent float64
		if previousTotal > 0 {
			changePercent = ((currentTotal - previousTotal) / previousTotal) * 100
		} else if currentTotal > 0 {
			changePercent = 100
		}

		result.CategoryChanges[cat] = CategoryChange{
			Name:         cat,
			CurrentTotal: currentTotal,
			PreviousTotal: previousTotal,
			ChangePercent: changePercent,
		}

		// Detect spikes (>20% increase)
		if changePercent > 20 && previousTotal > 0 {
			result.SpendingSpikes = append(result.SpendingSpikes, SpendingSpike{
				Category:     cat,
				Amount:       currentTotal,
				ChangePercent: changePercent,
				Period:       currentStart,
			})
		}
	}

	// Sort spikes by change percent
	sort.Slice(result.SpendingSpikes, func(i, j int) bool {
		return result.SpendingSpikes[i].ChangePercent > result.SpendingSpikes[j].ChangePercent
	})

	return result, nil
}

