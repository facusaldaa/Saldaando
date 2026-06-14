package llm

import (
	"botGastosPareja/internal/config"
	"botGastosPareja/internal/database"
	"context"
	"fmt"
	"log"
)

// FallbackProvider wraps a primary provider with a fallback
type FallbackProvider struct {
	primary   LLMProvider
	fallback  LLMProvider
	useFallback bool
}

// NewLLMClient creates a new LLM client with GLM as primary and HuggingFace as fallback
func NewLLMClient(cfg *config.Config) (LLMProvider, error) {
	if cfg.GLMAPIKey == "" {
		return nil, fmt.Errorf("GLM_API_KEY is required")
	}

	// Create GLM client
	glmClient, err := NewGLMClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create GLM client: %w", err)
	}

	// Create HuggingFace client if fallback is enabled
	var hfClient LLMProvider
	if cfg.GLMFallbackToHF {
		hfClient, err = NewHuggingFaceClient(cfg)
		if err != nil {
			log.Printf("Warning: Failed to create HuggingFace client: %v. Fallback disabled.", err)
			// Continue without fallback
		}
	}

	return &FallbackProvider{
		primary:    glmClient,
		fallback:   hfClient,
		useFallback: cfg.GLMFallbackToHF && hfClient != nil,
	}, nil
}

// ParseExpense attempts to parse expense using primary provider, falls back to HuggingFace on error
func (f *FallbackProvider) ParseExpense(ctx context.Context, message string, context ExpenseContext) (*ParsedExpense, error) {
	result, err := f.primary.ParseExpense(ctx, message, context)
	if err != nil && f.useFallback && f.fallback != nil {
		log.Printf("GLM parsing failed, trying HuggingFace fallback: %v", err)
		return f.fallback.ParseExpense(ctx, message, context)
	}
	return result, err
}

// ParseExpenseFromImage attempts to parse expense from image using primary provider
func (f *FallbackProvider) ParseExpenseFromImage(ctx context.Context, imageData []byte, context ExpenseContext) (*ParsedExpense, error) {
	result, err := f.primary.ParseExpenseFromImage(ctx, imageData, context)
	if err != nil && f.useFallback && f.fallback != nil {
		log.Printf("GLM vision parsing failed, trying HuggingFace fallback: %v", err)
		return f.fallback.ParseExpenseFromImage(ctx, imageData, context)
	}
	return result, err
}

// AnalyzeExpenses attempts analysis using primary provider, falls back to HuggingFace on error
func (f *FallbackProvider) AnalyzeExpenses(ctx context.Context, expenses []*database.Expense, lobby *database.Lobby, analysisType string) (*AnalysisInsights, error) {
	result, err := f.primary.AnalyzeExpenses(ctx, expenses, lobby, analysisType)
	if err != nil && f.useFallback && f.fallback != nil {
		log.Printf("GLM analysis failed, trying HuggingFace fallback: %v", err)
		return f.fallback.AnalyzeExpenses(ctx, expenses, lobby, analysisType)
	}
	return result, err
}

// HealthCheck checks health of primary provider
func (f *FallbackProvider) HealthCheck(ctx context.Context) error {
	return f.primary.HealthCheck(ctx)
}
