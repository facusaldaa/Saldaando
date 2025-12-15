package bot

import (
	"botGastosPareja/pkg/utils"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// registerInviteCommands registers invitation-related commands
func (h *Handler) registerInviteCommands() {
	h.router.RegisterCommand("invite", h.handleInvite)
	h.router.RegisterCommand("regenerate_invite", h.handleRegenerateInvite)
}

// handleInvite handles the /invite command to show invitation token
func (h *Handler) handleInvite(handler *Handler, message *tgbotapi.Message, args string) {
	userID := message.From.ID
	translator := handler.getTranslator(userID)

	// Get user's lobby for this specific chat (group/private)
	lobby, err := handler.getLobbyForMessage(message)
	if err != nil || lobby == nil {
		handler.sendTranslatedMessage(userID, message.Chat.ID, "error_lobby_not_found")
		return
	}

	if !lobby.InviteToken.Valid {
		handler.sendMessage(message.Chat.ID,
			translator.T("error_no_invite_token"))
		return
	}

	formattedToken := utils.FormatInviteToken(lobby.InviteToken.String)
	msg := translator.T("invite_token_display", formattedToken, formattedToken)
	handler.sendMessage(message.Chat.ID, msg)
}

// handleRegenerateInvite handles the /regenerate_invite command
func (h *Handler) handleRegenerateInvite(handler *Handler, message *tgbotapi.Message, args string) {
	userID := message.From.ID
	translator := handler.getTranslator(userID)

	// Get user's lobby for this specific chat (group/private)
	lobby, err := handler.getLobbyForMessage(message)
	if err != nil || lobby == nil {
		handler.sendTranslatedMessage(userID, message.Chat.ID, "error_lobby_not_found")
		return
	}

	// Check if user is the lobby creator
	if lobby.User1TelegramID != userID {
		handler.sendMessage(message.Chat.ID,
			translator.T("error_not_lobby_owner"))
		return
	}

	newToken, err := handler.lobbyService.RegenerateInviteToken(lobby.ID)
	if err != nil {
		handler.sendMessage(message.Chat.ID,
			fmt.Sprintf("❌ Error: %v", err))
		return
	}

	formattedToken := utils.FormatInviteToken(newToken)
	msg := translator.T("invite_token_regenerated", formattedToken, formattedToken)
	handler.sendMessage(message.Chat.ID, msg)
}
