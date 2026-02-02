package llm

import (
	"botGastosPareja/internal/config"
	"botGastosPareja/internal/database"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// GLMClient implements LLMProvider for GLM-4.7 API
type GLMClient struct {
	apiKey     string
	modelName  string
	baseURL    string
	temperature float64
	httpClient *http.Client
}

// GLMRequest represents a chat completion request
type GLMRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// GLMResponse represents a chat completion response
type GLMResponse struct {
	ID      string   `json:"id"`
	Choices []Choice `json:"choices"`
	Error   *GLMError `json:"error,omitempty"`
}

// Choice represents a response choice
type Choice struct {
	Message Message `json:"message"`
}

// GLMError represents an API error
type GLMError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

// NewGLMClient creates a new GLM client
func NewGLMClient(cfg *config.Config) (*GLMClient, error) {
	return &GLMClient{
		apiKey:     cfg.GLMAPIKey,
		modelName:  cfg.GLMModelName,
		baseURL:    cfg.GLMBaseURL,
		temperature: cfg.GLMTemperature,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// ParseExpense parses an expense from natural language
func (c *GLMClient) ParseExpense(ctx context.Context, message string, context ExpenseContext) (*ParsedExpense, error) {
	prompt := buildExpenseParsingPrompt(message, context)
	
	response, err := c.callAPI(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("GLM API call failed: %w", err)
	}

	// Parse JSON response
	var parsed ParsedExpense
	if err := json.Unmarshal([]byte(response), &parsed); err != nil {
		// Try to extract JSON from markdown code blocks
		response = extractJSONFromResponse(response)
		if err := json.Unmarshal([]byte(response), &parsed); err != nil {
			return nil, fmt.Errorf("failed to parse GLM response: %w", err)
		}
	}

	return &parsed, nil
}

// AnalyzeExpenses generates AI insights about expenses
func (c *GLMClient) AnalyzeExpenses(ctx context.Context, expenses []*database.Expense, lobby *database.Lobby, analysisType string) (*AnalysisInsights, error) {
	prompt := buildAnalysisPrompt(expenses, lobby, analysisType)
	
	response, err := c.callAPI(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("GLM API call failed: %w", err)
	}

	// Parse JSON response
	var insights AnalysisInsights
	if err := json.Unmarshal([]byte(response), &insights); err != nil {
		// Try to extract JSON from markdown code blocks
		response = extractJSONFromResponse(response)
		if err := json.Unmarshal([]byte(response), &insights); err != nil {
			return nil, fmt.Errorf("failed to parse GLM response: %w", err)
		}
	}

	return &insights, nil
}

// HealthCheck verifies the API is accessible
func (c *GLMClient) HealthCheck(ctx context.Context) error {
	testPrompt := "Say 'OK' if you can read this."
	_, err := c.callAPI(ctx, testPrompt)
	return err
}

// callAPI makes an HTTP request to GLM API with retry logic
func (c *GLMClient) callAPI(ctx context.Context, prompt string) (string, error) {
	url := fmt.Sprintf("%s/chat/completions", c.baseURL)

	reqBody := GLMRequest{
		Model: c.modelName,
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: c.temperature,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Retry logic: 3 attempts with exponential backoff
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(attempt) * time.Second
			log.Printf("Retrying GLM API call (attempt %d) after %v", attempt+1, backoff)
			time.Sleep(backoff)
		}

		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			return "", fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("HTTP request failed: %w", err)
			continue
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = fmt.Errorf("failed to read response: %w", err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			var apiErr GLMError
			if err := json.Unmarshal(body, &apiErr); err == nil && apiErr.Message != "" {
				lastErr = fmt.Errorf("GLM API error: %s", apiErr.Message)
			} else {
				lastErr = fmt.Errorf("GLM API error: status %d, body: %s", resp.StatusCode, string(body))
			}
			// Don't retry on 4xx errors (client errors)
			if resp.StatusCode >= 400 && resp.StatusCode < 500 {
				return "", lastErr
			}
			continue
		}

		var glmResp GLMResponse
		if err := json.Unmarshal(body, &glmResp); err != nil {
			lastErr = fmt.Errorf("failed to parse response: %w", err)
			continue
		}

		if glmResp.Error != nil {
			lastErr = fmt.Errorf("GLM API error: %s", glmResp.Error.Message)
			continue
		}

		if len(glmResp.Choices) == 0 {
			lastErr = fmt.Errorf("no choices in response")
			continue
		}

		return glmResp.Choices[0].Message.Content, nil
	}

	return "", lastErr
}

// extractJSONFromResponse extracts JSON from markdown code blocks or plain text
func extractJSONFromResponse(response string) string {
	// Remove markdown code blocks if present
	if len(response) > 3 && response[0:3] == "```" {
		// Find the closing ```
		endIdx := len(response) - 1
		for i := len(response) - 1; i >= 0; i-- {
			if i >= 2 && response[i-2:i+1] == "```" {
				endIdx = i - 2
				break
			}
		}
		// Find the first newline after opening ```
		startIdx := 3
		for i := 3; i < len(response); i++ {
			if response[i] == '\n' {
				startIdx = i + 1
				break
			}
		}
		if endIdx > startIdx {
			response = response[startIdx:endIdx]
		}
	}
	return response
}
