package service

import (
	"botGastosPareja/internal/database"
	"botGastosPareja/internal/service/llm"
	"context"
)

// LLMService provides a high-level interface for LLM operations
type LLMService interface {
	ParseExpenseFromMessage(ctx context.Context, message string, context llm.ExpenseContext) (*llm.ParsedExpense, error)
	GenerateAnalysisInsights(ctx context.Context, expenses []*database.Expense, lobby *database.Lobby, analysisType string) (*llm.AnalysisInsights, error)
}

// llmServiceWrapper wraps the LLM provider
type llmServiceWrapper struct {
	provider llm.LLMProvider
}

// NewLLMService creates a new LLM service
func NewLLMService(provider llm.LLMProvider) LLMService {
	return &llmServiceWrapper{provider: provider}
}

// ParseExpenseFromMessage parses an expense from a natural language message
func (s *llmServiceWrapper) ParseExpenseFromMessage(ctx context.Context, message string, context llm.ExpenseContext) (*llm.ParsedExpense, error) {
	return s.provider.ParseExpense(ctx, message, context)
}

// GenerateAnalysisInsights generates AI-powered insights
func (s *llmServiceWrapper) GenerateAnalysisInsights(ctx context.Context, expenses []*database.Expense, lobby *database.Lobby, analysisType string) (*llm.AnalysisInsights, error) {
	return s.provider.AnalyzeExpenses(ctx, expenses, lobby, analysisType)
}
