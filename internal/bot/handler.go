package bot

import (
	"botGastosPareja/internal/database"
	"botGastosPareja/internal/service"
	"botGastosPareja/pkg/i18n"
	"log"
	"regexp"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

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

// Services interface for dependency injection (if needed)
type Services struct {
	UserService  *service.UserService
	LobbyService *service.LobbyService
}

// NewHandler creates a new bot handler
func NewHandler(bot *tgbotapi.BotAPI, db *database.DB) *Handler {
	router := NewRouter()
	userService := service.NewUserService(db)
	lobbyService := service.NewLobbyService(db)
	paymentMethodService := service.NewPaymentMethodService(db)
	expenseService := service.NewExpenseService(db)
	settlementService := service.NewSettlementService(db, expenseService, lobbyService)
	analysisService := service.NewAnalysisService(db, expenseService)
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
	}
	handler.registerCommands()
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
	// Handle interactive flows (will be implemented later)
	// For now, just acknowledge
}

// handleCallbackQuery processes inline keyboard button presses
func (h *Handler) handleCallbackQuery(query *tgbotapi.CallbackQuery) {
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
