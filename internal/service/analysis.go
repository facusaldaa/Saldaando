package service

import (
	"botGastosPareja/internal/database"
	"botGastosPareja/pkg/utils"
	"context"
	"fmt"
	"log"
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
	llmService    LLMService // Optional, for AI-powered analysis
}

// NewAnalysisService creates a new analysis service
func NewAnalysisService(db *database.DB, expenseService *ExpenseService) *AnalysisService {
	return &AnalysisService{
		db:            db,
		expenseService: expenseService,
	}
}

// SetLLMService sets the LLM service for AI-powered analysis
func (s *AnalysisService) SetLLMService(llmService LLMService) {
	s.llmService = llmService
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

// AIAnalysisResult represents AI-enhanced analysis results
type AIAnalysisResult struct {
	*AnalysisResult
	AIInsights *AIAnalysisInsights
}

// AIAnalysisInsights represents AI-generated insights
type AIAnalysisInsights struct {
	Narrative          string
	Patterns           []AIPattern
	Anomalies          []AIAnomaly
	Recommendations    []AIRecommendation
	BehavioralInsights []AIBehavioralInsight
}

// AIPattern represents a detected pattern
type AIPattern struct {
	Type        string
	Description string
	Category    string
	Amount      float64
	Change      float64
}

// AIAnomaly represents a detected anomaly
type AIAnomaly struct {
	Type        string
	Description string
	ExpenseID   int64
	Severity    string
}

// AIRecommendation represents a recommendation
type AIRecommendation struct {
	Type        string
	Description string
	Action      string
	Priority    string
}

// AIBehavioralInsight represents a behavioral insight
type AIBehavioralInsight struct {
	Type        string
	Description string
	Data        map[string]interface{}
}

// AnalyzeWithAI performs analysis with AI-powered insights
func (s *AnalysisService) AnalyzeWithAI(lobbyID int64) (*AIAnalysisResult, error) {
	// Get basic statistical analysis
	basicResult, err := s.AnalyzeMonthly(lobbyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get basic analysis: %w", err)
	}

	// Get lobby info
	lobbyService := NewLobbyService(s.db)
	lobby, err := lobbyService.GetLobbyByID(lobbyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get lobby: %w", err)
	}
	if lobby == nil {
		return nil, fmt.Errorf("lobby not found")
	}

	// Get expenses for analysis
	now := time.Now()
	currentStart, currentEnd := utils.GetMonthStartEnd(now.Year(), now.Month())
	expenses, err := s.expenseService.GetExpensesByLobby(lobbyID, &currentStart, &currentEnd, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get expenses: %w", err)
	}

	// Get AI insights if LLM service is available
	var aiInsights *AIAnalysisInsights
	if s.llmService != nil {
		ctx := context.Background()
		insights, err := s.llmService.GenerateAnalysisInsights(ctx, expenses, lobby, "monthly")
		if err != nil {
			log.Printf("Failed to get AI insights: %v", err)
			// Continue without AI insights
		} else {
			// Convert LLM insights to service format
			aiInsights = &AIAnalysisInsights{
				Narrative: insights.Narrative,
				Patterns:  make([]AIPattern, 0, len(insights.Patterns)),
				Anomalies: make([]AIAnomaly, 0, len(insights.Anomalies)),
				Recommendations: make([]AIRecommendation, 0, len(insights.Recommendations)),
				BehavioralInsights: make([]AIBehavioralInsight, 0, len(insights.BehavioralInsights)),
			}

			for _, p := range insights.Patterns {
				aiInsights.Patterns = append(aiInsights.Patterns, AIPattern{
					Type:        p.Type,
					Description: p.Description,
					Category:    p.Category,
					Amount:      p.Amount,
					Change:      p.Change,
				})
			}

			for _, a := range insights.Anomalies {
				aiInsights.Anomalies = append(aiInsights.Anomalies, AIAnomaly{
					Type:        a.Type,
					Description: a.Description,
					ExpenseID:   a.ExpenseID,
					Severity:    a.Severity,
				})
			}

			for _, r := range insights.Recommendations {
				aiInsights.Recommendations = append(aiInsights.Recommendations, AIRecommendation{
					Type:        r.Type,
					Description: r.Description,
					Action:      r.Action,
					Priority:    r.Priority,
				})
			}

			for _, b := range insights.BehavioralInsights {
				aiInsights.BehavioralInsights = append(aiInsights.BehavioralInsights, AIBehavioralInsight{
					Type:        b.Type,
					Description: b.Description,
					Data:        b.Data,
				})
			}

			// Add summary stats if available
			if insights.SummaryStats != nil {
				// Summary stats are already in the basic result, but we can enhance with AI insights
			}
		}
	}

	return &AIAnalysisResult{
		AnalysisResult: basicResult,
		AIInsights:     aiInsights,
	}, nil
}

