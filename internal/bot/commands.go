package bot

import (
	"botGastosPareja/pkg/utils"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// registerCommands registers all bot commands
func (h *Handler) registerCommands() {
	// Basic commands
	h.router.RegisterCommand("start", h.handleStart)
	h.router.RegisterCommand("help", h.handleHelp)

	// Settings commands
	h.registerSettingsCommands()

	// Payment method commands
	h.registerPaymentMethodCommands()

	// Expense commands
	h.registerExpenseCommands()

	// Settlement commands
	h.registerSettlementCommands()

	// Reporting commands
	h.registerReportingCommands()

	// Analysis commands
	h.registerAnalysisCommands()

	// Language commands
	h.registerLanguageCommands()

	// Invite commands
	h.registerInviteCommands()
}

// handleStart handles the /start command
func (h *Handler) handleStart(handler *Handler, message *tgbotapi.Message, args string) {
	// In channels, message.From can be nil - we need to handle this
	if message.From == nil {
		handler.sendMessage(message.Chat.ID, "❌ Error: This command must be used by a user, not from a channel post.")
		return
	}

	userID := message.From.ID
	username := message.From.UserName
	displayName := message.From.FirstName
	if message.From.LastName != "" {
		displayName += " " + message.From.LastName
	}

	// Determine if this is a group/channel
	var groupChatID *int64
	isGroup := message.Chat.IsGroup() || message.Chat.IsSuperGroup() || message.Chat.IsChannel()
	if isGroup {
		groupID := message.Chat.ID
		groupChatID = &groupID
	}

	// Create or get user
	user, err := handler.userService.GetOrCreateUser(userID, username, displayName)
	if err != nil {
		// Use English for error message since we don't know user's language yet
		handler.sendMessage(message.Chat.ID, "❌ Error: Failed to initialize user. Please try again.")
		return
	}

	// Check if this is a new user (created within last 5 seconds)
	isNewUser := time.Since(user.CreatedAt) < 5*time.Second

	// Get translator with user's language preference
	translator := handler.getTranslator(userID)

	// Check if user is already in a lobby FOR THIS GROUP (or private)
	lobby, err := handler.lobbyService.GetLobbyByUserIDAndGroup(userID, groupChatID)
	if err != nil {
		handler.sendTranslatedMessage(userID, message.Chat.ID, "error_lobby_check")
		return
	}

	if lobby != nil {
		// User is already in a lobby for this group/private chat
		var partnerInfo string
		if lobby.User1TelegramID == userID {
			if lobby.User2TelegramID == 0 {
				partnerInfo = translator.T("waiting_partner")
			} else {
				partnerInfo = translator.T("partner_id", lobby.User2TelegramID)
			}
		} else {
			partnerInfo = translator.T("partner_id", lobby.User1TelegramID)
		}

		welcomeMsg := translator.T("welcome_back", displayName, lobby.ID, lobby.AccountType, partnerInfo)
		handler.sendMessage(message.Chat.ID, welcomeMsg)
		return
	}

	// For groups/channels: check if there's already a lobby for this group
	// If so, try to join it automatically
	if groupChatID != nil {
		existingLobby, err := handler.lobbyService.GetLobbyByGroupChatID(*groupChatID)
		if err == nil && existingLobby != nil {
			// There's already a lobby for this group
			// If it has space and user is not already in it, join automatically
			if existingLobby.User2TelegramID == 0 && existingLobby.User1TelegramID != userID {
				// Join the existing lobby
				err = handler.lobbyService.JoinLobbyDirectly(existingLobby.ID, userID)
				if err == nil {
					// Successfully joined - lobby is now complete with both users
					partnerInfo := translator.T("partner_id", existingLobby.User1TelegramID)
					welcomeMsg := translator.T("lobby_ready_group", displayName, existingLobby.ID, existingLobby.AccountType, partnerInfo)
					handler.sendMessage(message.Chat.ID, welcomeMsg)
					return
				}
				// If join failed, log and continue to create new lobby
			} else if existingLobby.User1TelegramID == userID || existingLobby.User2TelegramID == userID {
				// User is already in this lobby
				var partnerInfo string
				if existingLobby.User1TelegramID == userID {
					if existingLobby.User2TelegramID == 0 {
						partnerInfo = translator.T("waiting_partner")
					} else {
						partnerInfo = translator.T("partner_id", existingLobby.User2TelegramID)
					}
				} else {
					partnerInfo = translator.T("partner_id", existingLobby.User1TelegramID)
				}
				welcomeMsg := translator.T("welcome_back", displayName, existingLobby.ID, existingLobby.AccountType, partnerInfo)
				handler.sendMessage(message.Chat.ID, welcomeMsg)
				return
			}
			// Lobby is full, continue to create new one or show error
		}
	}

	// Check if user wants to join an existing lobby
	argsParts := parseCommandArgs(args)
	if len(argsParts) > 0 {
		// Try to join lobby by invitation token (pass groupChatID for validation)
		inviteToken := argsParts[0]
		err := handler.lobbyService.JoinLobbyByToken(inviteToken, userID, groupChatID)
		if err != nil {
			handler.sendTranslatedMessage(userID, message.Chat.ID, "error_lobby_join", err)
			return
		}

		handler.sendTranslatedMessage(userID, message.Chat.ID, "lobby_joined_token")
		return
	}

	// For new users, prompt language selection first (only in private chats)
	if isNewUser && !isGroup {
		handler.promptLanguageSelection(userID, message.Chat.ID)
		return
	}

	// Create new lobby for this group/private chat
	newLobby, err := handler.lobbyService.CreateLobby(userID, "separate", groupChatID)
	if err != nil {
		handler.sendTranslatedMessage(userID, message.Chat.ID, "error_lobby_create")
		return
	}

	// Show different messages for group vs private chat
	if isGroup {
		// In groups, partner can just run /start in the same group
		// For 2-person groups, the lobby is ready - just waiting for partner to run /start
		welcomeMsg := translator.T("lobby_created_group", displayName, newLobby.ID, newLobby.AccountType)
		handler.sendMessage(message.Chat.ID, welcomeMsg)
	} else {
		// In private chats, use token-based invitation
		formattedToken := utils.FormatInviteToken(newLobby.InviteToken.String)
		welcomeMsg := translator.T("lobby_created", displayName, newLobby.ID, newLobby.AccountType, formattedToken, formattedToken)
		handler.sendMessage(message.Chat.ID, welcomeMsg)

		// Send security instructions (token appears twice in the message)
		securityMsg := translator.T("lobby_security_info", formattedToken, formattedToken)
		handler.sendMessage(message.Chat.ID, securityMsg)
	}
}

// handleHelp handles the /help command
func (h *Handler) handleHelp(handler *Handler, message *tgbotapi.Message, args string) {
	if message.From == nil {
		handler.sendMessage(message.Chat.ID, "❌ Error: This command must be used by a user.")
		return
	}
	userID := message.From.ID
	translator := handler.getTranslator(userID)
	helpText := translator.T("help")
	handler.sendMessage(message.Chat.ID, helpText)
}

// promptLanguageSelection prompts a new user to select their language
func (h *Handler) promptLanguageSelection(userID int64, chatID int64) {
	// Use English for the prompt since user hasn't selected language yet
	msg := "🌐 *Select your language / Selecciona tu idioma:*\n\n" +
		"Please choose your preferred language to continue.\n" +
		"Por favor elige tu idioma preferido para continuar."

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🇬🇧 English", "lang_en"),
			tgbotapi.NewInlineKeyboardButtonData("🇦🇷 Español", "lang_es_AR"),
		),
	)

	h.sendMessageWithKeyboard(chatID, msg, keyboard)
}

// parseCommandArgs parses command arguments into parts
func parseCommandArgs(args string) []string {
	if args == "" {
		return []string{}
	}
	return strings.Fields(args)
}
