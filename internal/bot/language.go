package bot

import (
	"botGastosPareja/pkg/i18n"
	"botGastosPareja/pkg/utils"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// registerLanguageCommands registers language-related commands
func (h *Handler) registerLanguageCommands() {
	h.router.RegisterCommand("language", h.handleLanguage)
	h.router.RegisterCallback("lang_en", h.handleLanguageCallback)
	h.router.RegisterCallback("lang_es_AR", h.handleLanguageCallback)
}

// handleLanguage handles the /language command
func (h *Handler) handleLanguage(handler *Handler, message *tgbotapi.Message, args string) {
	userID := message.From.ID
	translator := handler.getTranslator(userID)

	argsParts := parseCommandArgs(args)
	if len(argsParts) == 0 {
		// Show current language and available languages
		currentLang := translator.GetLanguage()
		langName := getLanguageName(currentLang)

		availableLangs := "• en - English\n• es_AR - Español (Argentina)"

		msg := translator.T("language_current", langName, availableLangs)
		handler.sendMessage(message.Chat.ID, msg)
		return
	}

	// Change language
	langCode := strings.ToLower(argsParts[0])
	var newLang i18n.Language

	switch langCode {
	case "en", "english":
		newLang = i18n.LanguageEnglish
	case "es_ar", "es", "spanish", "español":
		newLang = i18n.LanguageSpanishAR
	default:
		availableLangs := "en, es_AR"
		msg := translator.T("language_invalid", availableLangs)
		handler.sendMessage(message.Chat.ID, msg)
		return
	}

	// Update user's language preference
	err := handler.userService.UpdateUserLanguage(userID, newLang)
	if err != nil {
		handler.sendMessage(message.Chat.ID,
			fmt.Sprintf("❌ Error: Failed to update language: %v", err))
		return
	}

	// Get new translator for confirmation message
	newTranslator := i18n.NewTranslator(newLang)
	langName := getLanguageName(newLang)
	msg := newTranslator.T("language_changed", langName)
	handler.sendMessage(message.Chat.ID, msg)
}

// handleLanguageCallback handles language selection from inline keyboard
func (h *Handler) handleLanguageCallback(handler *Handler, query *tgbotapi.CallbackQuery) {
	userID := query.From.ID
	langCode := query.Data // "lang_en" or "lang_es_AR"

	var newLang i18n.Language
	var langName string

	switch langCode {
	case "lang_en":
		newLang = i18n.LanguageEnglish
		langName = "English"
	case "lang_es_AR":
		newLang = i18n.LanguageSpanishAR
		langName = "Español (Argentina)"
	default:
		// Acknowledge callback
		callback := tgbotapi.NewCallback(query.ID, "Invalid language selection")
		handler.bot.Request(callback)
		return
	}

	// Update user's language preference
	err := handler.userService.UpdateUserLanguage(userID, newLang)
	if err != nil {
		callback := tgbotapi.NewCallback(query.ID, "❌ Error updating language")
		handler.bot.Request(callback)
		return
	}

	// Acknowledge callback
	callback := tgbotapi.NewCallback(query.ID, fmt.Sprintf("✅ Language set to %s", langName))
	handler.bot.Request(callback)

	// Get new translator for confirmation message
	newTranslator := i18n.NewTranslator(newLang)
	msg := newTranslator.T("language_changed", langName)

	// Edit the message to remove keyboard
	editMsg := tgbotapi.NewEditMessageText(query.Message.Chat.ID, query.Message.MessageID, msg)
	editMsg.ParseMode = tgbotapi.ModeHTML
	// Convert markdown to HTML
	msg = convertMarkdownToHTML(msg)
	editMsg.Text = msg
	handler.bot.Send(editMsg)

	// Continue with lobby creation after language selection
	handler.continueStartAfterLanguage(userID, query.Message.Chat.ID, query.From.FirstName, query.From.LastName)
}

// continueStartAfterLanguage continues the /start flow after language selection
func (h *Handler) continueStartAfterLanguage(userID int64, chatID int64, firstName, lastName string) {
	displayName := firstName
	if lastName != "" {
		displayName += " " + lastName
	}

	translator := h.getTranslator(userID)

	// Note: For language selection callback, we don't have the message context
	// So we use the regular lookup - this is OK since language is user-specific, not group-specific
	// Check if user is already in a lobby
	lobby, err := h.lobbyService.GetLobbyByUserID(userID)
	if err != nil {
		h.sendTranslatedMessage(userID, chatID, "error_lobby_check")
		return
	}

	if lobby != nil {
		// User is already in a lobby
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
		h.sendMessage(chatID, welcomeMsg)
		return
	}

	// Create new lobby (for private chats only)
	newLobby, err := h.lobbyService.CreateLobby(userID, "separate", nil)
	if err != nil {
		h.sendTranslatedMessage(userID, chatID, "error_lobby_create")
		return
	}

	// Format invitation token for display
	formattedToken := utils.FormatInviteToken(newLobby.InviteToken.String)
	welcomeMsg := translator.T("lobby_created", displayName, newLobby.ID, newLobby.AccountType, formattedToken, formattedToken)
	h.sendMessage(chatID, welcomeMsg)

	// Send security instructions (token appears twice in the message)
	securityMsg := translator.T("lobby_security_info", formattedToken, formattedToken)
	h.sendMessage(chatID, securityMsg)
}

// getLanguageName returns the display name for a language
func getLanguageName(lang i18n.Language) string {
	switch lang {
	case i18n.LanguageSpanishAR:
		return "Español (Argentina)"
	case i18n.LanguageEnglish:
		return "English"
	default:
		return string(lang)
	}
}
