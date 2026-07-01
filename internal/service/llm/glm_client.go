package llm

import (
	"botGastosPareja/internal/config"
	"botGastosPareja/internal/database"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// GLMClient implements LLMProvider for GLM-4 API
type GLMClient struct {
	apiKey      string
	modelName   string
	visionModel string // separate model for vision (e.g., "glm-4v-plus"); empty = use modelName
	baseURL     string
	temperature float64
	httpClient  *http.Client
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

// VisionContentPart represents a single part in a multi-modal message
type VisionContentPart struct {
	Type     string           `json:"type"`
	Text     string           `json:"text,omitempty"`
	ImageURL *VisionImageURL  `json:"image_url,omitempty"`
}

// VisionImageURL represents an image URL in a vision request
type VisionImageURL struct {
	URL string `json:"url"`
}

// VisionRequest represents a multi-modal chat completion request (text + image)
type VisionRequest struct {
	Model       string            `json:"model"`
	Messages    []VisionMessage   `json:"messages"`
	Temperature float64           `json:"temperature,omitempty"`
}

// VisionMessage represents a message with multi-modal content
type VisionMessage struct {
	Role    string              `json:"role"`
	Content []VisionContentPart `json:"content"`
}

// NewGLMClient creates a new GLM client
func NewGLMClient(cfg *config.Config) (*GLMClient, error) {
	return &GLMClient{
		apiKey:      cfg.GLMAPIKey,
		modelName:   cfg.GLMModelName,
		visionModel: cfg.GLMVisionModel,
		baseURL:     cfg.GLMBaseURL,
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

// ParseExpenseFromImage parses an expense from a receipt image using vision
func (c *GLMClient) ParseExpenseFromImage(ctx context.Context, imageData []byte, context ExpenseContext) (*ParsedExpense, error) {
	// Use vision model if set, otherwise fall back to main model
	visionModel := c.visionModel
	if visionModel == "" {
		visionModel = c.modelName
	}

	// Build a receipt OCR prompt
	prompt := buildReceiptOCRPrompt(context)

	// Encode image to base64 data URI
	encoded := base64.StdEncoding.EncodeToString(imageData)
	mimeType := detectImageMimeType(imageData)
	dataURI := fmt.Sprintf("data:%s;base64,%s", mimeType, encoded)

	response, err := c.callVisionAPI(ctx, prompt, dataURI, visionModel)
	if err != nil {
		return nil, fmt.Errorf("GLM vision API call failed: %w", err)
	}

	// Parse JSON response
	var parsed ParsedExpense
	if err := json.Unmarshal([]byte(response), &parsed); err != nil {
		// Try to extract JSON from markdown code blocks
		response = extractJSONFromResponse(response)
		if err := json.Unmarshal([]byte(response), &parsed); err != nil {
			return nil, fmt.Errorf("failed to parse GLM vision response: %w", err)
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
			// Don't retry on 4xx errors (client errors), except 429 (rate limit)
			if resp.StatusCode >= 400 && resp.StatusCode < 500 && resp.StatusCode != 429 {
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

// callVisionAPI sends a multi-modal request (text + image) to the GLM API
func (c *GLMClient) callVisionAPI(ctx context.Context, textPrompt string, imageDataURI string, model string) (string, error) {
	url := fmt.Sprintf("%s/chat/completions", c.baseURL)

	reqBody := VisionRequest{
		Model: model,
		Messages: []VisionMessage{
			{
				Role: "user",
				Content: []VisionContentPart{
					{Type: "text", Text: textPrompt},
					{Type: "image_url", ImageURL: &VisionImageURL{URL: imageDataURI}},
				},
			},
		},
		Temperature: c.temperature,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal vision request: %w", err)
	}

	// Retry logic: 3 attempts with exponential backoff
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(attempt) * time.Second
			log.Printf("Retrying GLM vision API call (attempt %d) after %v", attempt+1, backoff)
			time.Sleep(backoff)
		}

		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			return "", fmt.Errorf("failed to create vision request: %w", err)
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
			lastErr = fmt.Errorf("failed to read vision response: %w", err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			var apiErr GLMError
			if err := json.Unmarshal(body, &apiErr); err == nil && apiErr.Message != "" {
				lastErr = fmt.Errorf("GLM vision API error: %s", apiErr.Message)
			} else {
				lastErr = fmt.Errorf("GLM vision API error: status %d, body: %s", resp.StatusCode, string(body))
			}
			if resp.StatusCode >= 400 && resp.StatusCode < 500 {
				return "", lastErr
			}
			continue
		}

		var glmResp GLMResponse
		if err := json.Unmarshal(body, &glmResp); err != nil {
			lastErr = fmt.Errorf("failed to parse vision response: %w", err)
			continue
		}

		if glmResp.Error != nil {
			lastErr = fmt.Errorf("GLM vision API error: %s", glmResp.Error.Message)
			continue
		}

		if len(glmResp.Choices) == 0 {
			lastErr = fmt.Errorf("no choices in vision response")
			continue
		}

		return glmResp.Choices[0].Message.Content, nil
	}

	return "", lastErr
}

// buildReceiptOCRPrompt builds a prompt for receipt OCR
func buildReceiptOCRPrompt(context ExpenseContext) string {
	var prompt strings.Builder

	if context.UserLanguage == "es_AR" {
		prompt.WriteString(`Sos un asistente experto en leer recibos y facturas en argentino.`)
		prompt.WriteString("\n\n")
		prompt.WriteString("# INSTRUCCIONES\n")
		prompt.WriteString("Mirá la imagen del recibo/factura/ticket y extraé la información del gasto.\n")
		prompt.WriteString("Buscá el MONTO TOTAL, una DESCRIPCIÓN del comercio/producto, y la FECHA.\n")
		prompt.WriteString("Si no encontrás una categoría clara, inferila del contexto:\n")
		prompt.WriteString("  - supermercado, chino, almacén, carrefour, coto, dia, makro → comida\n")
		prompt.WriteString("  - restaurante, bar, mc, mcdonald's, burger, mostaza → comida\n")
		prompt.WriteString("  - uber, taxi, cabify, didi, rapipago → transporte\n")
		prompt.WriteString("  - farmacia, farma, farmacity → salud\n")
		prompt.WriteString("  - cine, netflix, spotify, steam → entretenimiento\n")
		prompt.WriteString("  - frávega, ml, mercadolibre, amazon → otros\n")
		prompt.WriteString("\n")
	} else {
		prompt.WriteString(`You are an expert assistant at reading receipts and invoices.`)
		prompt.WriteString("\n\n")
		prompt.WriteString("# INSTRUCTIONS\n")
		prompt.WriteString("Look at the receipt/invoice/ticket image and extract the expense information.\n")
		prompt.WriteString("Find the TOTAL AMOUNT, a DESCRIPTION of the store/product, and the DATE.\n")
		prompt.WriteString("If no clear category is found, infer it from context:\n")
		prompt.WriteString("  - supermarket, grocery, walmart, target, costco → food\n")
		prompt.WriteString("  - restaurant, bar, mcdonald's, starbucks → food\n")
		prompt.WriteString("  - uber, taxi, lyft, gas station → transport\n")
		prompt.WriteString("  - pharmacy, cvs, walgreens → health\n")
		prompt.WriteString("  - cinema, netflix, spotify, steam → entertainment\n")
		prompt.WriteString("  - amazon, bestbuy, walmart online → other\n")
		prompt.WriteString("\n")
	}

	// Available categories
	if len(context.Categories) > 0 {
		if context.UserLanguage == "es_AR" {
			prompt.WriteString("Categorías disponibles:\n")
		} else {
			prompt.WriteString("Available categories:\n")
		}
		for _, cat := range context.Categories {
			prompt.WriteString(fmt.Sprintf("  - %s\n", cat))
		}
		prompt.WriteString("\n")
	}

	// Payment methods
	if len(context.PaymentMethods) > 0 {
		if context.UserLanguage == "es_AR" {
			prompt.WriteString("Métodos de pago disponibles:\n")
		} else {
			prompt.WriteString("Available payment methods:\n")
		}
		for _, pm := range context.PaymentMethods {
			if pm.IsActive {
				prompt.WriteString(fmt.Sprintf("  - %s\n", pm.Name))
			}
		}
		prompt.WriteString("\n")
	}

	// Output format
	if context.UserLanguage == "es_AR" {
		prompt.WriteString("# FORMATO DE SALIDA REQUERIDO\n")
		prompt.WriteString("Devolvé ÚNICAMENTE un objeto JSON válido (sin markdown, sin explicaciones):\n\n")
	} else {
		prompt.WriteString("# REQUIRED OUTPUT FORMAT\n")
		prompt.WriteString("Return ONLY a valid JSON object (no markdown, no explanations):\n\n")
	}
	prompt.WriteString(`{"amount": <number>, "description": "<string>", "category": "<string>", "payment_method": "", "spender": "me", "date": "YYYY-MM-DD|today", "confidence": <0.0-1.0>, "needs_clarification": false, "clarification_question": "", "is_installment": false, "installments": 0}`)

	return prompt.String()
}

// detectImageMimeType detects the MIME type from image bytes
func detectImageMimeType(data []byte) string {
	if len(data) == 0 {
		return "image/jpeg"
	}
	// Check magic bytes
	if len(data) > 3 && data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		return "image/jpeg"
	}
	if len(data) > 7 && data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
		return "image/png"
	}
	if len(data) > 3 && data[0] == 0x47 && data[1] == 0x49 && data[2] == 0x46 {
		return "image/gif"
	}
	if len(data) > 3 && data[0] == 0x52 && data[1] == 0x49 && data[2] == 0x46 && data[3] == 0x46 {
		return "image/webp"
	}
	return "image/jpeg" // Default to JPEG
}
