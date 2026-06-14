package llm

import (
	"botGastosPareja/internal/database"
	"context"
)

// ExpenseContext contains context information for parsing expenses
type ExpenseContext struct {
	Lobby            *database.Lobby
	PaymentMethods   []*database.PaymentMethod
	Categories       []string
	RecentExpenses   []*database.Expense
	UserLanguage     string
	UserID           int64
	PartnerName      string
	UserDisplayName  string
}

// ParsedExpense represents a parsed expense from natural language
type ParsedExpense struct {
	Amount              float64 `json:"amount"`
	Description         string  `json:"description"`
	Category            string  `json:"category"`
	PaymentMethod       string  `json:"payment_method"`
	Spender             string  `json:"spender"` // "me", "partner", "user1", "user2"
	Date                string  `json:"date"`    // ISO date or "today", "yesterday", etc.
	Confidence          float64 `json:"confidence"`
	NeedsClarification  bool    `json:"needs_clarification"`
	ClarificationQuestion string `json:"clarification_question"`
	IsInstallment       bool    `json:"is_installment"` // true if this is an installment purchase
	Installments        int     `json:"installments"`   // number of installments (cuotas)
}

// LLMProvider defines the interface for LLM providers
type LLMProvider interface {
	ParseExpense(ctx context.Context, message string, context ExpenseContext) (*ParsedExpense, error)
	ParseExpenseFromImage(ctx context.Context, imageData []byte, context ExpenseContext) (*ParsedExpense, error)
	AnalyzeExpenses(ctx context.Context, expenses []*database.Expense, lobby *database.Lobby, analysisType string) (*AnalysisInsights, error)
	HealthCheck(ctx context.Context) error
}

// AnalysisInsights represents AI-generated insights about expenses
type AnalysisInsights struct {
	Narrative           string
	Patterns            []Pattern
	Anomalies           []Anomaly
	Recommendations     []Recommendation
	BehavioralInsights  []BehavioralInsight
	SummaryStats        *SummaryStats
}

// SummaryStats represents summary statistics for analysis
type SummaryStats struct {
	TotalExpenses    float64
	ExpenseCount     int
	AverageExpense   float64
	User1Percentage  float64
	User2Percentage  float64
	TopCategory      string
	MonthlyTrend     string // "increasing", "decreasing", "stable"
}

// Pattern represents a detected spending pattern
type Pattern struct {
	Type        string // "trend", "category_change", "velocity", "seasonal"
	Description string
	Category    string
	Amount      float64
	Change      float64 // percentage change
	TimePeriod  string
}

// Anomaly represents a detected anomaly
type Anomaly struct {
	Type        string // "unusual_amount", "duplicate", "temporal", "missing_data"
	Description string
	ExpenseID   int64
	Severity    string // "low", "medium", "high"
	Suggestion  string
}

// Recommendation represents an actionable recommendation
type Recommendation struct {
	Type           string // "budget", "category", "timing", "balance"
	Description    string
	Action         string
	Priority       string // "low", "medium", "high"
	ExpectedImpact string
}

// BehavioralInsight represents a behavioral insight
type BehavioralInsight struct {
	Type        string // "velocity", "partner_comparison", "category_mix", "temporal_pattern"
	Description string
	Data        map[string]interface{}
}
