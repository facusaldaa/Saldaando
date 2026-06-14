package llm

import (
	"botGastosPareja/internal/config"
	"botGastosPareja/internal/database"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HuggingFaceClient implements LLMProvider for HuggingFace Inference API
type HuggingFaceClient struct {
	apiKey     string
	modelName  string
	httpClient *http.Client
}

// HFRequest represents a HuggingFace API request
type HFRequest struct {
	Inputs     string                 `json:"inputs"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// HFResponse represents a HuggingFace API response
type HFResponse struct {
	GeneratedText string `json:"generated_text"`
	Error         string `json:"error,omitempty"`
}

// NewHuggingFaceClient creates a new HuggingFace client
func NewHuggingFaceClient(cfg *config.Config) (*HuggingFaceClient, error) {
	return &HuggingFaceClient{
		apiKey:     cfg.HuggingFaceAPIKey,
		modelName:  cfg.HuggingFaceModel,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// ParseExpense parses an expense from natural language
func (c *HuggingFaceClient) ParseExpense(ctx context.Context, message string, context ExpenseContext) (*ParsedExpense, error) {
	prompt := buildExpenseParsingPrompt(message, context)
	
	response, err := c.callAPI(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("HuggingFace API call failed: %w", err)
	}

	// Parse JSON response
	var parsed ParsedExpense
	if err := json.Unmarshal([]byte(response), &parsed); err != nil {
		// Try to extract JSON from markdown code blocks
		response = extractJSONFromResponse(response)
		if err := json.Unmarshal([]byte(response), &parsed); err != nil {
			return nil, fmt.Errorf("failed to parse HuggingFace response: %w", err)
		}
	}

	return &parsed, nil
}

// ParseExpenseFromImage returns error — HuggingFace does not support vision
func (c *HuggingFaceClient) ParseExpenseFromImage(ctx context.Context, imageData []byte, context ExpenseContext) (*ParsedExpense, error) {
	return nil, fmt.Errorf("HuggingFace does not support image/vision processing")
}

// AnalyzeExpenses generates AI insights about expenses
func (c *HuggingFaceClient) AnalyzeExpenses(ctx context.Context, expenses []*database.Expense, lobby *database.Lobby, analysisType string) (*AnalysisInsights, error) {
	prompt := buildAnalysisPrompt(expenses, lobby, analysisType)
	
	response, err := c.callAPI(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("HuggingFace API call failed: %w", err)
	}

	// Parse JSON response
	var insights AnalysisInsights
	if err := json.Unmarshal([]byte(response), &insights); err != nil {
		// Try to extract JSON from markdown code blocks
		response = extractJSONFromResponse(response)
		if err := json.Unmarshal([]byte(response), &insights); err != nil {
			return nil, fmt.Errorf("failed to parse HuggingFace response: %w", err)
		}
	}

	return &insights, nil
}

// HealthCheck verifies the API is accessible
func (c *HuggingFaceClient) HealthCheck(ctx context.Context) error {
	testPrompt := "Say 'OK' if you can read this."
	_, err := c.callAPI(ctx, testPrompt)
	return err
}

// callAPI makes an HTTP request to HuggingFace Inference API
func (c *HuggingFaceClient) callAPI(ctx context.Context, prompt string) (string, error) {
	url := fmt.Sprintf("https://api-inference.huggingface.co/models/%s", c.modelName)

	reqBody := HFRequest{
		Inputs: prompt,
		Parameters: map[string]interface{}{
			"max_new_tokens": 500,
			"temperature":     0.3,
			"return_full_text": false,
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var apiErr struct {
			Error string `json:"error"`
		}
		if err := json.Unmarshal(body, &apiErr); err == nil && apiErr.Error != "" {
			return "", fmt.Errorf("HuggingFace API error: %s", apiErr.Error)
		}
		return "", fmt.Errorf("HuggingFace API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	// HuggingFace can return either a single object or an array
	var hfResp []HFResponse
	if err := json.Unmarshal(body, &hfResp); err != nil {
		// Try single object
		var singleResp HFResponse
		if err := json.Unmarshal(body, &singleResp); err != nil {
			return "", fmt.Errorf("failed to parse response: %w", err)
		}
		if singleResp.Error != "" {
			return "", fmt.Errorf("HuggingFace API error: %s", singleResp.Error)
		}
		return singleResp.GeneratedText, nil
	}

	if len(hfResp) == 0 {
		return "", fmt.Errorf("empty response from HuggingFace")
	}

	if hfResp[0].Error != "" {
		return "", fmt.Errorf("HuggingFace API error: %s", hfResp[0].Error)
	}

	return hfResp[0].GeneratedText, nil
}

// (uses extractJSONFromResponse from glm_client.go)
