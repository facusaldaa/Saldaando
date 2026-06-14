package bot

import (
	"botGastosPareja/pkg/i18n"
	"botGastosPareja/pkg/utils"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// MaxImageSize is the maximum allowed image size in bytes (20MB)
const MaxImageSize = 20 * 1024 * 1024

// MaxPDFPages is the max number of PDF pages to process (first N pages)
const MaxPDFPages = 3

// handlePhoto processes a photo message — OCR via vision API → confirmation
func (h *Handler) handlePhoto(message *tgbotapi.Message) {
	log.Printf("handlePhoto called: chatID=%d, From=nil?=%v, Photo=%d, Document=%v",
		message.Chat.ID, message.From == nil, len(message.Photo), message.Document != nil)

	if message.From == nil {
		log.Printf("handlePhoto: cannot process — message has no From field")
		return
	}

	userID := message.From.ID

	// Must have LLM enabled for OCR
	if !h.llmEnabled || h.llmService == nil {
		h.sendTranslatedMessage(userID, message.Chat.ID, "receipt_llm_disabled")
		return
	}

	// Must be in a lobby
	lobby, err := h.getLobbyForMessage(message)
	if err != nil || lobby == nil {
		h.sendTranslatedMessage(userID, message.Chat.ID, "error_lobby_not_found")
		return
	}

	// Rate limiting: max 1 OCR per 10 seconds per user
	if lastTime, exists := h.rateLimiter[userID]; exists {
		if time.Since(lastTime) < 10*time.Second {
			return // Rate limited — silently ignore
		}
	}
	h.rateLimiter[userID] = time.Now()

	translator := h.getTranslator(userID)
	isSpanish := translator.GetLanguage() == i18n.LanguageSpanishAR

	// Send "processing" message
	processingMsg := "⏳"
	if isSpanish {
		processingMsg = "⏳ Analizando recibo..."
	} else {
		processingMsg = "⏳ Analyzing receipt..."
	}
	statusMsg := tgbotapi.NewMessage(message.Chat.ID, processingMsg)
	sent, _ := h.bot.Send(statusMsg)

	// Download the photo (pick largest available)
	imageData, err := h.downloadPhoto(message)
	if err != nil {
		log.Printf("Failed to download photo: %v", err)
		editMsg := tgbotapi.NewEditMessageText(message.Chat.ID, sent.MessageID, "❌ Error al descargar la imagen.")
		h.bot.Send(editMsg)
		return
	}

	// Build expense context
	expenseCtx, err := h.buildExpenseContext(lobby, userID)
	if err != nil {
		log.Printf("Failed to build expense context: %v", err)
		editMsg := tgbotapi.NewEditMessageText(message.Chat.ID, sent.MessageID, "❌ Error interno.")
		h.bot.Send(editMsg)
		return
	}

	// Call vision API
	log.Printf("Calling vision OCR for user %d, lobby %d (image: %d bytes)", userID, lobby.ID, len(imageData))
	parsed, err := h.llmService.ParseExpenseFromImage(context.Background(), imageData, expenseCtx)
	if err != nil {
		log.Printf("Vision OCR failed: %v", err)
		if isSpanish {
			editMsg := tgbotapi.NewEditMessageText(message.Chat.ID, sent.MessageID,
				"❌ No pude leer el recibo. Asegurate de que la foto sea clara y tenga el monto visible.\n\nPodés usar /add para agregar el gasto manualmente.")
			h.bot.Send(editMsg)
		} else {
			editMsg := tgbotapi.NewEditMessageText(message.Chat.ID, sent.MessageID,
				"❌ Couldn't read the receipt. Make sure the photo is clear and the amount is visible.\n\nYou can use /add to add the expense manually.")
			h.bot.Send(editMsg)
		}
		return
	}

	// Validate parsed amount
	if parsed.Amount <= 0 {
		if isSpanish {
			editMsg := tgbotapi.NewEditMessageText(message.Chat.ID, sent.MessageID,
				"❌ No encontré un monto válido en el recibo. Podés usar /add para agregarlo manualmente.")
			h.bot.Send(editMsg)
		} else {
			editMsg := tgbotapi.NewEditMessageText(message.Chat.ID, sent.MessageID,
				"❌ Couldn't find a valid amount in the receipt. Use /add to add it manually.")
			h.bot.Send(editMsg)
		}
		return
	}

	// Delete the "processing" message
	deleteMsg := tgbotapi.NewDeleteMessage(message.Chat.ID, sent.MessageID)
	h.bot.Send(deleteMsg)

	// Normalize category
	if parsed.Category != "" {
		if canonical, ok := utils.NormalizeCategory(parsed.Category); ok {
			parsed.Category = canonical
		}
	}

	// Parse date
	expenseDate := time.Now()
	if parsed.Date != "" {
		if parsed.Date == "today" || parsed.Date == "hoy" {
			expenseDate = time.Now()
		} else if parsed.Date == "yesterday" || parsed.Date == "ayer" {
			expenseDate = time.Now().AddDate(0, 0, -1)
		} else {
			if parsedDate, err := utils.ParseDate(parsed.Date); err == nil {
				expenseDate = parsedDate
			}
		}
	}

	// Find payment method if specified
	var paymentMethodID *int64
	if parsed.PaymentMethod != "" {
		methods, err := h.paymentMethodService.GetPaymentMethodsByLobby(lobby.ID, true)
		if err == nil {
			for _, method := range methods {
				if method.Name == parsed.PaymentMethod {
					paymentMethodID = &method.ID
					break
				}
			}
		}
	}

	// Store pending expense for confirmation (same flow as NL parsing)
	spenderID := userID
	if parsed.Spender != "" {
		spenderLower := toLower(parsed.Spender)
		if spenderLower == "partner" || spenderLower == "pareja" || spenderLower == "user2" {
			if lobby.User2TelegramID != 0 {
				spenderID = lobby.User2TelegramID
			}
		}
	}

	pending := &PendingExpense{
		Parsed:            parsed,
		LobbyID:           lobby.ID,
		SpenderID:         spenderID,
		ExpenseDate:       expenseDate,
		PaymentMethodID:   paymentMethodID,
		OriginalMessageID: message.MessageID,
		ChatID:            message.Chat.ID,
	}
	h.pendingExpenses[userID] = pending

	// Show confirmation using the existing flow
	h.showExpenseConfirmation(userID, message.Chat.ID, pending, translator)
}

// downloadPhoto downloads the largest photo from a message, or converts a PDF to image
func (h *Handler) downloadPhoto(message *tgbotapi.Message) ([]byte, error) {
	var fileID string
	var fileSize int
	isPDF := false

	if len(message.Photo) > 0 {
		// Pick the largest photo
		largest := message.Photo[len(message.Photo)-1]
		fileID = largest.FileID
		fileSize = largest.FileSize
	} else if message.Document != nil {
		mime := message.Document.MimeType
		if mime == "image/jpeg" || mime == "image/png" || mime == "image/webp" {
			fileID = message.Document.FileID
			fileSize = message.Document.FileSize
		} else if mime == "application/pdf" {
			fileID = message.Document.FileID
			fileSize = message.Document.FileSize
			isPDF = true
		} else {
			return nil, fmt.Errorf("unsupported document type: %s", mime)
		}
	} else {
		return nil, fmt.Errorf("no photo or document found")
	}

	if fileSize > MaxImageSize {
		return nil, fmt.Errorf("file too large: %d bytes (max %d)", fileSize, MaxImageSize)
	}

	// Get file path from Telegram
	fileConfig := tgbotapi.FileConfig{FileID: fileID}
	file, err := h.bot.GetFile(fileConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to get file from Telegram: %w", err)
	}

	// Download the file
	fileURL := file.Link(h.bot.Token)
	resp, err := http.Get(fileURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("file download returned status %d", resp.StatusCode)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read file data: %w", err)
	}

	// Convert PDF to PNG image using pdftoppm
	if isPDF {
		imgData, err := convertPDFToImage(buf.Bytes())
		if err != nil {
			return nil, fmt.Errorf("failed to convert PDF to image: %w", err)
		}
		return imgData, nil
	}

	return buf.Bytes(), nil
}

// convertPDFToImage converts the first page of a PDF to a PNG image using pdftoppm
func convertPDFToImage(pdfData []byte) ([]byte, error) {
	tmpDir, err := os.MkdirTemp("", "pdfocr-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write PDF to temp file
	pdfPath := tmpDir + "/input.pdf"
	if err := os.WriteFile(pdfPath, pdfData, 0644); err != nil {
		return nil, fmt.Errorf("failed to write temp PDF: %w", err)
	}

	// Convert first page to PNG
	outPrefix := tmpDir + "/page"
	cmd := exec.Command("pdftoppm", "-png", "-r", "200", "-f", "1", "-l", "1", pdfPath, outPrefix)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("pdftoppm failed: %s: %w", string(output), err)
	}

	// Read the generated PNG
	imgPath := outPrefix + "-1.png"
	imgData, err := os.ReadFile(imgPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read PDF page image: %w", err)
	}

	return imgData, nil
}

// toLower is a simple ASCII toLower helper for spender matching
func toLower(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 32
		}
		b[i] = c
	}
	return string(b)
}

// getDocMime returns the MIME type of a document, or empty string if nil
func getDocMime(doc *tgbotapi.Document) string {
	if doc == nil {
		return ""
	}
	return doc.MimeType
}
