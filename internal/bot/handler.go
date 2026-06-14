package bot

import (
	"botGastosPareja/internal/database"
	"botGastosPareja/internal/service"
	"botGastosPareja/internal/service/llm"
	"botGastosPareja/pkg/i18n"
	"botGastosPareja/pkg/utils"
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// PendingExpense stores a parsed expense waiting for user confirmation
type PendingExpense struct {
	Parsed            *llm.ParsedExpense
	LobbyID           int64
	SpenderID         int64
	ExpenseDate       time.Time
	PaymentMethodID   *int64
	OriginalMessageID int
	ChatID            int64
}

// Handler handles Telegram bot updates
type Handler struct {
	bot                  *tgbotapi.BotAPI
	db                   *database.DB
	router               *Router
	userService          *service.UserService
	lobbyService         *service.LobbyService
	paymentMethodService *service.PaymentMethodService
	expenseService       *service.ExpenseService
	settlementService    *service.SettlementService
	analysisService      *service.AnalysisService
	llmService           service.LLMService
	rateLimiter          map[int64]time.Time       // User ID -> last NL parse time
	pendingExpenses      map[int64]*PendingExpense // User ID -> pending expense
	dynamicMsgs          *DynamicMessageTracker    // tracks messages for in-place editing
	llmEnabled           bool                      // Feature flag for NL and AI features
	cuotasEnabled        bool                      // Feature flag for /add_cuotas
}

// getTranslator gets a translator for a user
func (h *Handler) getTranslator(userID int64) *i18n.Translator {
	user, err := h.userService.GetOrCreateUser(userID, "", "")
	if err != nil || user == nil {
		return i18n.NewTranslator(i18n.LanguageEnglish) // Default to English
	}

	if user.Language.Valid {
		return i18n.NewTranslator(i18n.Language(user.Language.String))
	}
	return i18n.NewTranslator(i18n.LanguageEnglish) // Default to English
}

// getUserDisplayName gets a user's display name, falling back to username or Telegram ID.
// It never returns "User 1" / "User 2" — always uses real info or their Telegram ID.
func (h *Handler) getUserDisplayName(telegramID int64) string {
	user, err := h.userService.GetUserByTelegramID(telegramID)
	if err != nil || user == nil {
		return fmt.Sprintf("User %d", telegramID)
	}
	if user.DisplayName.Valid && user.DisplayName.String != "" {
		return user.DisplayName.String
	}
	if user.Username.Valid && user.Username.String != "" {
		return "@" + user.Username.String
	}
	return fmt.Sprintf("User %d", telegramID)
}

// Services interface for dependency injection (if needed)
type Services struct {
	UserService  *service.UserService
	LobbyService *service.LobbyService
}

// NewHandler creates a new bot handler
func NewHandler(bot *tgbotapi.BotAPI, db *database.DB, llmService service.LLMService, llmEnabled bool, cuotasEnabled bool) *Handler {
	router := NewRouter()
	userService := service.NewUserService(db)
	lobbyService := service.NewLobbyService(db)
	paymentMethodService := service.NewPaymentMethodService(db)
	expenseService := service.NewExpenseService(db)
	settlementService := service.NewSettlementService(db, expenseService, lobbyService)
	analysisService := service.NewAnalysisService(db, expenseService)

	// Set LLM service for analysis if available
	if llmService != nil && llmEnabled {
		analysisService.SetLLMService(llmService)
	}

	handler := &Handler{
		bot:                  bot,
		db:                   db,
		router:               router,
		userService:          userService,
		lobbyService:         lobbyService,
		paymentMethodService: paymentMethodService,
		expenseService:       expenseService,
		settlementService:    settlementService,
		analysisService:      analysisService,
		llmService:           llmService,
		rateLimiter:          make(map[int64]time.Time),
		pendingExpenses:      make(map[int64]*PendingExpense),
		dynamicMsgs:          NewDynamicMessageTracker(),
		llmEnabled:           llmEnabled,
		cuotasEnabled:        cuotasEnabled,
	}
	handler.registerCommands()
	handler.registerCallbacks()
	return handler
}

// RegisterCommands registers all bot commands
func (h *Handler) RegisterCommands() {
	h.registerCommands()
}

// RegisterTelegramCommands registers bot commands with Telegram API
func (h *Handler) RegisterTelegramCommands() error {
	commands := []tgbotapi.BotCommand{
		{
			Command:     "start",
			Description: "Start the bot and create/join a lobby",
		},
		{
			Command:     "help",
			Description: "Show available commands",
		},
		{
			Command:     "add",
			Description: "Add an expense",
		},
		{
			Command:     "list",
			Description: "List expenses",
		},
		{
			Command:     "summary",
			Description: "Get expense summary",
		},
		{
			Command:     "settle",
			Description: "Calculate who owes whom",
		},
		{
			Command:     "review",
			Description: "Get AI-powered expense review",
		},
		{
			Command:     "settings",
			Description: "Configure lobby settings",
		},
		{
			Command:     "language",
			Description: "Change language / Cambiar idioma",
		},
		{
			Command:     "payment_methods",
			Description: "Manage payment methods",
		},
	}

	cmd := tgbotapi.NewSetMyCommands(commands...)
	_, err := h.bot.Request(cmd)
	return err
}

// HandleUpdate processes incoming Telegram updates
func (h *Handler) HandleUpdate(update tgbotapi.Update) {
	// Log all updates for debugging
	log.Printf("Update received: UpdateID=%d, CallbackQuery=%v, Message=%v, ChannelPost=%v, EditedChannelPost=%v",
		update.UpdateID,
		update.CallbackQuery != nil,
		update.Message != nil,
		update.ChannelPost != nil,
		update.EditedChannelPost != nil)

	// Handle callback queries (inline keyboard buttons) first
	if update.CallbackQuery != nil {
		h.handleCallbackQuery(update.CallbackQuery)
		return
	}

	// Handle channel posts
	// Channel posts can be from the channel itself (From == nil) or from users (From != nil)
	if update.ChannelPost != nil {
		log.Printf("Received channel post: ChatID=%d, Text=%s, From=%v",
			update.ChannelPost.Chat.ID,
			update.ChannelPost.Text,
			update.ChannelPost.From)

		// If channel post has a From field, it's from a user - process it as a regular message
		if update.ChannelPost.From != nil {
			// Process channel post from user as a regular message
			if update.ChannelPost.IsCommand() {
				h.handleCommand(update.ChannelPost)
				return
			}
			h.handleMessage(update.ChannelPost)
			return
		}
		// Channel posts without From are from the channel itself - can't process commands
		return
	}

	// Handle edited channel posts
	if update.EditedChannelPost != nil {
		log.Printf("Received edited channel post: ChatID=%d", update.EditedChannelPost.Chat.ID)
		return
	}

	// Handle regular messages (from users in groups/channels/private chats)
	if update.Message == nil {
		log.Printf("Update has no message, callback, or channel post - ignoring")
		return
	}

	log.Printf("Received message: ChatID=%d, ChatType=%s, From=%v, Text=%s, IsCommand=%v",
		update.Message.Chat.ID,
		getChatType(update.Message.Chat),
		update.Message.From,
		update.Message.Text,
		update.Message.IsCommand())

	// Handle commands
	if update.Message.IsCommand() {
		h.handleCommand(update.Message)
		return
	}

	// Handle photo messages (receipt OCR)
	hasPhoto := update.Message.Photo != nil && len(update.Message.Photo) > 0
	hasImageDoc := update.Message.Document != nil && (update.Message.Document.MimeType == "image/jpeg" || update.Message.Document.MimeType == "image/png" || update.Message.Document.MimeType == "image/webp")

	log.Printf("DEBUG msg Parse: Text=%q Caption=%q Photo=%d Doc=%v", update.Message.Text, update.Message.Caption, len(update.Message.Photo), update.Message.Document != nil)

	if hasPhoto || hasImageDoc {
		h.handlePhoto(update.Message)
		return
	}

	// Handle image documents
	if update.Message.Document != nil {
		mime := update.Message.Document.MimeType
		if mime == "image/jpeg" || mime == "image/png" || mime == "image/webp" {
			h.handlePhoto(update.Message)
			return
		}
	}

	// Handle regular messages (for interactive flows)
	h.handleMessage(update.Message)
}

// getChatType returns a string representation of the chat type
func getChatType(chat *tgbotapi.Chat) string {
	if chat.IsChannel() {
		return "channel"
	}
	if chat.IsSuperGroup() {
		return "supergroup"
	}
	if chat.IsGroup() {
		return "group"
	}
	return "private"
}

// handleCommand processes bot commands
func (h *Handler) handleCommand(message *tgbotapi.Message) {
	// In channels, message.From can be nil for channel posts - skip those
	// But regular messages in channels/groups should have From set
	if message.From == nil {
		log.Printf("Skipping command from message without From field: ChatID=%d, Text=%s",
			message.Chat.ID, message.Text)
		return
	}

	log.Printf("Processing command: UserID=%d, ChatID=%d, Command=%s, Args=%s",
		message.From.ID, message.Chat.ID, message.Command(), message.CommandArguments())

	command := message.Command()
	args := message.CommandArguments()

	handler := h.router.GetCommandHandler(command)
	if handler != nil {
		handler(h, message, args)
	} else {
		userID := message.From.ID
		h.sendTranslatedMessage(userID, message.Chat.ID, "error_unknown_command")
	}
}

// handleMessage processes regular text messages
func (h *Handler) handleMessage(message *tgbotapi.Message) {
	if message.From == nil || message.Text == "" {
		return
	}

	userID := message.From.ID
	translator := h.getTranslator(userID)

	// Check if user is in a lobby
	lobby, err := h.getLobbyForMessage(message)
	if err != nil || lobby == nil {
		// Not in a lobby, ignore message
		return
	}

	// Rate limiting: max 1 NL parse per 5 seconds per user
	if lastTime, exists := h.rateLimiter[userID]; exists {
		if time.Since(lastTime) < 5*time.Second {
			return // Rate limited
		}
	}
	h.rateLimiter[userID] = time.Now()

	// Attempt natural language parsing
	if h.llmEnabled && h.llmService != nil {
		h.handleNaturalLanguageExpense(message, lobby, translator)
	}
}

// handleNaturalLanguageExpense attempts to parse and create expense from natural language
func (h *Handler) handleNaturalLanguageExpense(message *tgbotapi.Message, lobby *database.Lobby, translator *i18n.Translator) {
	userID := message.From.ID
	ctx := context.Background()

	// Build expense context
	expenseContext, err := h.buildExpenseContext(lobby, userID)
	if err != nil {
		log.Printf("Failed to build expense context: %v", err)
		return
	}

	// Parse expense using LLM
	parsed, err := h.llmService.ParseExpenseFromMessage(ctx, message.Text, expenseContext)
	if err != nil {
		log.Printf("LLM parsing failed: %v", err)
		// Fallback: show /add usage
		h.sendTranslatedMessage(userID, message.Chat.ID, "expense_add_usage")
		return
	}

	// Check if clarification is needed
	if parsed.NeedsClarification && parsed.ClarificationQuestion != "" {
		h.sendMessage(message.Chat.ID, parsed.ClarificationQuestion)
		return
	}

	// Validate parsed expense
	if parsed.Amount <= 0 {
		h.sendTranslatedMessage(userID, message.Chat.ID, "expense_invalid_amount")
		return
	}

	// Normalize category
	if parsed.Category != "" {
		if canonical, ok := utils.NormalizeCategory(parsed.Category); ok {
			parsed.Category = canonical
		}
	}

	// Validate installments require payment method
	if parsed.IsInstallment && parsed.Installments > 0 {
		if parsed.PaymentMethod == "" {
			if expenseContext.UserLanguage == "es_AR" {
				h.sendMessage(message.Chat.ID, "❌ Para agregar gastos en cuotas, necesitás especificar un método de pago (tarjeta de crédito).")
			} else {
				h.sendMessage(message.Chat.ID, "❌ To add installment expenses, you need to specify a payment method (credit card).")
			}
			return
		}
	}

	// Map spender
	spenderID := userID // Default to the user adding the expense
	if parsed.Spender != "" {
		spenderLower := strings.ToLower(parsed.Spender)
		if spenderLower == "partner" || spenderLower == "pareja" || spenderLower == "user2" {
			if lobby.User2TelegramID != 0 {
				spenderID = lobby.User2TelegramID
			} else {
				h.sendTranslatedMessage(userID, message.Chat.ID, "waiting_partner")
				return
			}
		} else if spenderLower == "user1" {
			spenderID = lobby.User1TelegramID
		} else if spenderLower == "me" || spenderLower == "yo" {
			spenderID = userID
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
			// Try to parse as date
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
				if strings.EqualFold(method.Name, parsed.PaymentMethod) {
					paymentMethodID = &method.ID
					break
				}
			}
		}
	}

	// Store pending expense for confirmation
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

	// Show confirmation message with inline keyboard
	h.showExpenseConfirmation(userID, message.Chat.ID, pending, translator)
}

// showExpenseConfirmation shows a confirmation message with parsed expense details
func (h *Handler) showExpenseConfirmation(userID int64, chatID int64, pending *PendingExpense, translator *i18n.Translator) {
	var msg strings.Builder

	// Detect language directly from translator settings
	isSpanish := translator.GetLanguage() == i18n.LanguageSpanishAR

	if isSpanish {
		msg.WriteString("📝 *Gasto detectado:*\n\n")
	} else {
		msg.WriteString("📝 *Expense detected:*\n\n")
	}

	// Amount
	if isSpanish {
		msg.WriteString(fmt.Sprintf("💰 *Monto:* %s\n", utils.FormatCurrency(pending.Parsed.Amount)))
	} else {
		msg.WriteString(fmt.Sprintf("💰 *Amount:* %s\n", utils.FormatCurrency(pending.Parsed.Amount)))
	}

	// Description
	if pending.Parsed.Description != "" {
		if isSpanish {
			msg.WriteString(fmt.Sprintf("📦 *Descripción:* %s\n", pending.Parsed.Description))
		} else {
			msg.WriteString(fmt.Sprintf("📦 *Description:* %s\n", pending.Parsed.Description))
		}
	}

	// Category
	if pending.Parsed.Category != "" {
		if isSpanish {
			msg.WriteString(fmt.Sprintf("🏷️ *Categoría:* %s\n", pending.Parsed.Category))
		} else {
			msg.WriteString(fmt.Sprintf("🏷️ *Category:* %s\n", pending.Parsed.Category))
		}
	}

	// Payment method
	if pending.Parsed.PaymentMethod != "" {
		if isSpanish {
			msg.WriteString(fmt.Sprintf("💳 *Método de pago:* %s\n", pending.Parsed.PaymentMethod))
		} else {
			msg.WriteString(fmt.Sprintf("💳 *Payment method:* %s\n", pending.Parsed.PaymentMethod))
		}
	}

	// Spender
	spenderText := ""
	if isSpanish {
		if pending.Parsed.Spender == "me" || pending.Parsed.Spender == "yo" {
			spenderText = "Vos"
		} else if pending.Parsed.Spender == "partner" || pending.Parsed.Spender == "pareja" {
			spenderText = "Tu pareja"
		} else {
			spenderText = pending.Parsed.Spender
		}
		msg.WriteString(fmt.Sprintf("👤 *Pagador:* %s\n", spenderText))
	} else {
		if pending.Parsed.Spender == "me" {
			spenderText = "You"
		} else if pending.Parsed.Spender == "partner" {
			spenderText = "Your partner"
		} else {
			spenderText = pending.Parsed.Spender
		}
		msg.WriteString(fmt.Sprintf("👤 *Spender:* %s\n", spenderText))
	}

	// Date
	dateText := utils.FormatDate(pending.ExpenseDate)
	if pending.Parsed.Date == "today" || pending.Parsed.Date == "hoy" {
		if isSpanish {
			dateText = "Hoy"
		} else {
			dateText = "Today"
		}
	} else if pending.Parsed.Date == "yesterday" || pending.Parsed.Date == "ayer" {
		if isSpanish {
			dateText = "Ayer"
		} else {
			dateText = "Yesterday"
		}
	}
	if isSpanish {
		msg.WriteString(fmt.Sprintf("📅 *Fecha:* %s\n", dateText))
	} else {
		msg.WriteString(fmt.Sprintf("📅 *Date:* %s\n", dateText))
	}

	// Installments
	if pending.Parsed.IsInstallment && pending.Parsed.Installments > 0 {
		if isSpanish {
			msg.WriteString(fmt.Sprintf("\n📊 *Cuotas:* %d", pending.Parsed.Installments))
		} else {
			msg.WriteString(fmt.Sprintf("\n📊 *Installments:* %d", pending.Parsed.Installments))
		}
	}

	// Confidence
	if isSpanish {
		msg.WriteString(fmt.Sprintf("\n*Confianza:* %.0f%%", pending.Parsed.Confidence*100))
	} else {
		msg.WriteString(fmt.Sprintf("\n*Confidence:* %.0f%%", pending.Parsed.Confidence*100))
	}

	// Create inline keyboard
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Confirmar", fmt.Sprintf("exp_confirm_%d", userID)),
			tgbotapi.NewInlineKeyboardButtonData("✏️ Editar", fmt.Sprintf("exp_edit_%d", userID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ Cancelar", fmt.Sprintf("exp_cancel_%d", userID)),
		),
	)

	h.sendMessageWithKeyboard(chatID, msg.String(), keyboard)
}

// confirmExpense creates the expense after user confirmation.
// It edits the original confirmation message in-place instead of sending a new one.
func (h *Handler) confirmExpense(userID int64) error {
	pending, exists := h.pendingExpenses[userID]
	if !exists {
		return fmt.Errorf("no pending expense found")
	}

	translator := h.getTranslator(userID)
	var msg string
	isSpanish := translator.GetLanguage() == i18n.LanguageSpanishAR

	// Check if this is an installment purchase
	if pending.Parsed.IsInstallment && pending.Parsed.Installments > 0 {
		// Create installments
		if pending.PaymentMethodID == nil {
			return fmt.Errorf("payment method required for installments")
		}

		expenses, err := h.expenseService.CreateInstallmentExpenses(
			pending.LobbyID,
			pending.SpenderID,
			pending.Parsed.Amount,
			pending.Parsed.Description,
			pending.Parsed.Category,
			pending.ExpenseDate,
			pending.PaymentMethodID,
			int64(pending.Parsed.Installments),
		)
		if err != nil {
			return fmt.Errorf("failed to create installments: %w", err)
		}

		// Remove from pending
		delete(h.pendingExpenses, userID)

		// Build confirmation text
		if isSpanish {
			msg = "✅ *Gasto en cuotas registrado:*\n\n"
		} else {
			msg = "✅ *Installment expense recorded:*\n\n"
		}
		msg += translator.T("installments_created",
			utils.FormatCurrency(pending.Parsed.Amount),
			pending.Parsed.Description,
			pending.Parsed.Installments)
		msg += fmt.Sprintf("\n")
		for i, exp := range expenses {
			msg += fmt.Sprintf("• Cuota %d/%d: ID %d - %s (%s)\n",
				i+1, len(expenses), exp.ID,
				utils.FormatCurrency(exp.Amount),
				utils.FormatDate(exp.ExpenseDate))
		}
	} else {
		// Create regular expense
		expense, err := h.expenseService.CreateExpense(
			pending.LobbyID,
			pending.SpenderID,
			pending.Parsed.Amount,
			pending.Parsed.Description,
			pending.Parsed.Category,
			pending.ExpenseDate,
			pending.PaymentMethodID,
		)
		if err != nil {
			return fmt.Errorf("failed to create expense: %w", err)
		}

		// Remove from pending
		delete(h.pendingExpenses, userID)

		// Build confirmation text
		if isSpanish {
			msg = "✅ *Gasto registrado:*\n\n"
		} else {
			msg = "✅ *Expense recorded:*\n\n"
		}
		msg += translator.T("expense_added",
			utils.FormatCurrency(expense.Amount),
			expense.Description.String)
		msg += fmt.Sprintf("ID: %d\n", expense.ID)

		if expense.Category.Valid {
			msg += translator.T("expense_category", expense.Category.String)
		}
		if expense.PaymentMethodID.Valid {
			pm, _ := h.paymentMethodService.GetPaymentMethodByID(expense.PaymentMethodID.Int64)
			if pm != nil {
				msg += translator.T("expense_payment_method", pm.Name)
			}
		}
		msg += translator.T("expense_list_date", utils.FormatDate(expense.ExpenseDate))
	}

	// Edit the original confirmation message instead of sending a new one
	edit := tgbotapi.NewEditMessageText(pending.ChatID, pending.OriginalMessageID, convertMarkdownToHTML(msg))
	edit.ParseMode = tgbotapi.ModeHTML
	if _, err := h.bot.Send(edit); err != nil {
		log.Printf("Error editing confirmation message: %v — sending new", err)
		h.sendMessage(pending.ChatID, msg)
	}

	return nil
}

// cancelPendingExpense cancels a pending expense.
// It edits the original confirmation message in-place instead of sending a new one.
func (h *Handler) cancelPendingExpense(userID int64) {
	pending, exists := h.pendingExpenses[userID]
	if !exists {
		return
	}

	delete(h.pendingExpenses, userID)

	translator := h.getTranslator(userID)
	// Check if Spanish
	isSpanish := translator.GetLanguage() == i18n.LanguageSpanishAR

	var msg string
	if isSpanish {
		msg = "❌ *Gasto cancelado. No se guardó nada.*"
	} else {
		msg = "❌ *Expense cancelled. Nothing was saved.*"
	}

	// Edit the original confirmation message instead of sending a new one
	edit := tgbotapi.NewEditMessageText(pending.ChatID, pending.OriginalMessageID, convertMarkdownToHTML(msg))
	edit.ParseMode = tgbotapi.ModeHTML
	if _, err := h.bot.Send(edit); err != nil {
		log.Printf("Error editing cancellation message: %v — sending new", err)
		h.sendMessage(pending.ChatID, msg)
	}
}

// buildExpenseContext builds context for LLM expense parsing
func (h *Handler) buildExpenseContext(lobby *database.Lobby, userID int64) (llm.ExpenseContext, error) {
	ctx := llm.ExpenseContext{
		Lobby:  lobby,
		UserID: userID,
	}

	// Get user info
	user, err := h.userService.GetUserByTelegramID(userID)
	if err == nil && user != nil {
		if user.Language.Valid {
			ctx.UserLanguage = user.Language.String
		}
		if user.DisplayName.Valid {
			ctx.UserDisplayName = user.DisplayName.String
		} else if user.Username.Valid {
			ctx.UserDisplayName = "@" + user.Username.String
		}
	}

	// Get partner info
	var partnerID int64
	if lobby.User1TelegramID == userID {
		partnerID = lobby.User2TelegramID
	} else {
		partnerID = lobby.User1TelegramID
	}
	if partnerID != 0 {
		partner, err := h.userService.GetUserByTelegramID(partnerID)
		if err == nil && partner != nil {
			if partner.DisplayName.Valid {
				ctx.PartnerName = partner.DisplayName.String
			} else if partner.Username.Valid {
				ctx.PartnerName = "@" + partner.Username.String
			}
		}
	}

	// Get payment methods
	paymentMethods, err := h.paymentMethodService.GetPaymentMethodsByLobby(lobby.ID, true)
	if err == nil {
		ctx.PaymentMethods = paymentMethods
	}

	// Get recent expenses (last 5)
	now := time.Now()
	startDate := now.AddDate(0, 0, -30) // Last 30 days
	expenses, err := h.expenseService.GetExpensesByLobby(lobby.ID, &startDate, &now, nil)
	if err == nil && len(expenses) > 0 {
		if len(expenses) > 5 {
			ctx.RecentExpenses = expenses[:5]
		} else {
			ctx.RecentExpenses = expenses
		}
	}

	// Extract categories from recent expenses
	categoryMap := make(map[string]bool)
	for _, exp := range expenses {
		if exp.Category.Valid {
			categoryMap[exp.Category.String] = true
		}
	}
	ctx.Categories = make([]string, 0, len(categoryMap))
	for cat := range categoryMap {
		ctx.Categories = append(ctx.Categories, cat)
	}

	return ctx, nil
}

// handleCallbackQuery processes inline keyboard button presses
func (h *Handler) handleCallbackQuery(query *tgbotapi.CallbackQuery) {
	// Handle expense confirmation callbacks
	if strings.HasPrefix(query.Data, "exp_confirm_") {
		userID, err := extractUserIDFromCallback(query.Data, "exp_confirm_")
		if err == nil && query.From.ID == userID {
			err := h.confirmExpense(userID)
			if err != nil {
				log.Printf("Error confirming expense: %v", err)
				h.sendMessage(query.Message.Chat.ID, "❌ Error al confirmar el gasto. Por favor intentá de nuevo.")
			}
		}
		// Acknowledge callback
		callback := tgbotapi.NewCallback(query.ID, "")
		h.bot.Request(callback)
		return
	}

	if strings.HasPrefix(query.Data, "exp_cancel_") {
		userID, err := extractUserIDFromCallback(query.Data, "exp_cancel_")
		if err == nil && query.From.ID == userID {
			h.cancelPendingExpense(userID)
		}
		// Acknowledge callback
		callback := tgbotapi.NewCallback(query.ID, "")
		h.bot.Request(callback)
		return
	}

	if strings.HasPrefix(query.Data, "exp_edit_") {
		userID, err := extractUserIDFromCallback(query.Data, "exp_edit_")
		if err == nil && query.From.ID == userID {
			translator := h.getTranslator(userID)
			testMsg := translator.T("waiting_partner")
			isSpanish := strings.Contains(testMsg, "pareja") || strings.Contains(testMsg, "Esperando")

			var msg string
			if isSpanish {
				msg = "✏️ Para editar, usá el comando /add con los valores correctos.\n\nEjemplo:\n/add 5000 supermercado comida"
			} else {
				msg = "✏️ To edit, use the /add command with the correct values.\n\nExample:\n/add 5000 supermarket food"
			}
			h.sendMessage(query.Message.Chat.ID, msg)
			h.cancelPendingExpense(userID)
		}
		// Acknowledge callback
		callback := tgbotapi.NewCallback(query.ID, "")
		h.bot.Request(callback)
		return
	}

	// Try router for other callbacks
	handler := h.router.GetCallbackHandler(query.Data)
	if handler != nil {
		handler(h, query)
	} else {
		// Acknowledge callback
		callback := tgbotapi.NewCallback(query.ID, "")
		if _, err := h.bot.Request(callback); err != nil {
			log.Printf("Error acknowledging callback: %v", err)
		}
	}
}

// extractUserIDFromCallback extracts user ID from callback data
func extractUserIDFromCallback(callbackData, prefix string) (int64, error) {
	userIDStr := strings.TrimPrefix(callbackData, prefix)
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid user ID in callback: %w", err)
	}
	return userID, nil
}

// registerCallbacks registers callback handlers
func (h *Handler) registerCallbacks() {
	// Callbacks are handled directly in handleCallbackQuery
	// This function is here for future extensibility
}

// sendMessage sends a text message to a chat
func (h *Handler) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	// Use HTML mode instead of Markdown to avoid parsing issues with special characters
	msg.ParseMode = tgbotapi.ModeHTML
	// Convert markdown-style formatting to HTML
	text = convertMarkdownToHTML(text)
	msg.Text = text
	if _, err := h.bot.Send(msg); err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

// convertMarkdownToHTML converts simple markdown to HTML for Telegram
func convertMarkdownToHTML(text string) string {
	// Escape HTML special characters first (but do this carefully to not double-escape)
	// We need to escape & first, then others
	text = strings.ReplaceAll(text, "&", "&amp;")
	text = strings.ReplaceAll(text, "<", "&lt;")
	text = strings.ReplaceAll(text, ">", "&gt;")

	// Convert `code` to <code>code</code> (do this before bold to avoid conflicts)
	codeRegex := regexp.MustCompile("`([^`]+)`")
	text = codeRegex.ReplaceAllString(text, "<code>$1</code>")

	// Convert *bold* to <b>bold</b>
	// Match *text* but not **text** or *text*text* (single asterisks for bold)
	// This regex matches * followed by non-asterisk chars followed by *
	boldRegex := regexp.MustCompile(`\*([^*\n]+)\*`)
	text = boldRegex.ReplaceAllString(text, "<b>$1</b>")

	return text
}

// sendTranslatedMessage sends a translated message to a user
func (h *Handler) sendTranslatedMessage(userID int64, chatID int64, key string, args ...interface{}) {
	translator := h.getTranslator(userID)
	text := translator.T(key, args...)
	h.sendMessage(chatID, text)
}

// getLobbyForMessage gets the lobby for a user in the context of the message's chat (group/private)
func (h *Handler) getLobbyForMessage(message *tgbotapi.Message) (*database.Lobby, error) {
	userID := message.From.ID

	// Determine if this is a group/channel
	var groupChatID *int64
	if message.Chat.IsGroup() || message.Chat.IsSuperGroup() || message.Chat.IsChannel() {
		groupID := message.Chat.ID
		groupChatID = &groupID
		log.Printf("DEBUG getLobbyForMessage: userID=%d, ChatID=%d, IsGroup=%v, IsSuperGroup=%v, IsChannel=%v",
			userID, groupID, message.Chat.IsGroup(), message.Chat.IsSuperGroup(), message.Chat.IsChannel())
	} else {
		log.Printf("DEBUG getLobbyForMessage: userID=%d, ChatID=%d, Private chat", userID, message.Chat.ID)
	}

	return h.lobbyService.GetLobbyByUserIDAndGroup(userID, groupChatID)
}

// sendMessageWithKeyboard sends a message with inline keyboard
func (h *Handler) sendMessageWithKeyboard(chatID int64, text string, keyboard tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeHTML
	// Convert markdown-style formatting to HTML
	text = convertMarkdownToHTML(text)
	msg.Text = text
	msg.ReplyMarkup = keyboard
	if _, err := h.bot.Send(msg); err != nil {
		log.Printf("Error sending message with keyboard: %v", err)
	}
}

// Getter methods for services (used by scheduler)
func (h *Handler) GetLobbyService() *service.LobbyService {
	return h.lobbyService
}

func (h *Handler) GetSettlementService() *service.SettlementService {
	return h.settlementService
}

func (h *Handler) GetUserService() *service.UserService {
	return h.userService
}

func (h *Handler) GetExpenseService() *service.ExpenseService {
	return h.expenseService
}
