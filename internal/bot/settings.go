package bot

import (
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// registerSettingsCommands registers settings-related commands
func (h *Handler) registerSettingsCommands() {
	h.router.RegisterCommand("settings", h.handleSettings)
}

// handleSettings handles the /settings command
func (h *Handler) handleSettings(handler *Handler, message *tgbotapi.Message, args string) {
	userID := message.From.ID
	translator := handler.getTranslator(userID)

	// Get user's lobby for this specific chat (group/private)
	lobby, err := handler.getLobbyForMessage(message)
	if err != nil {
		handler.sendTranslatedMessage(userID, message.Chat.ID, "error_generic", err)
		return
	}

	if lobby == nil {
		handler.sendTranslatedMessage(userID, message.Chat.ID, "error_lobby_not_found")
		return
	}

	argsParts := parseCommandArgs(args)
	if len(argsParts) == 0 {
		// Show current settings
		settingsMsg := translator.T("settings_current",
			lobby.ID,
			lobby.AccountType,
			lobby.User1SalaryPercentage*100,
			lobby.User2SalaryPercentage*100)
		handler.sendMessage(message.Chat.ID, settingsMsg)
		return
	}

	// Parse settings update
	settingType := strings.ToLower(argsParts[0])
	var accountType *string
	var user1Pct, user2Pct *float64

	switch settingType {
	case "account_type", "accounttype":
		if len(argsParts) < 2 {
			handler.sendTranslatedMessage(userID, message.Chat.ID, "settings_usage")
			return
		}
		at := strings.ToLower(argsParts[1])
		if at != "separate" && at != "shared" {
			handler.sendTranslatedMessage(userID, message.Chat.ID, "settings_invalid_type")
			return
		}
		accountType = &at

	case "salary":
		if len(argsParts) < 3 {
			handler.sendTranslatedMessage(userID, message.Chat.ID, "settings_salary_usage")
			return
		}
		pct1, err1 := strconv.ParseFloat(argsParts[1], 64)
		pct2, err2 := strconv.ParseFloat(argsParts[2], 64)
		if err1 != nil || err2 != nil {
			handler.sendTranslatedMessage(userID, message.Chat.ID, "settings_invalid_pct")
			return
		}
		if pct1 < 0 || pct1 > 1 || pct2 < 0 || pct2 > 1 {
			handler.sendTranslatedMessage(userID, message.Chat.ID, "settings_pct_range")
			return
		}
		user1Pct = &pct1
		user2Pct = &pct2

	default:
		handler.sendTranslatedMessage(userID, message.Chat.ID, "settings_unknown")
		return
	}

	// Update settings
	err = handler.lobbyService.UpdateLobbySettings(lobby.ID, accountType, user1Pct, user2Pct)
	if err != nil {
		handler.sendTranslatedMessage(userID, message.Chat.ID, "settings_error", err)
		return
	}

	handler.sendTranslatedMessage(userID, message.Chat.ID, "settings_updated")
}
