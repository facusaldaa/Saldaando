package bot

import (
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// registerPaymentMethodCommands registers payment method commands
func (h *Handler) registerPaymentMethodCommands() {
	h.router.RegisterCommand("payment_methods", h.handlePaymentMethods)
}

// handlePaymentMethods handles the /payment_methods command
func (h *Handler) handlePaymentMethods(handler *Handler, message *tgbotapi.Message, args string) {
	if message.From == nil {
		handler.sendMessage(message.Chat.ID, "❌ Error: This command must be used by a user.")
		return
	}
	userID := message.From.ID
	translator := handler.getTranslator(userID)

	// Get user's lobby for this specific chat (group/private)
	lobby, err := handler.getLobbyForMessage(message)
	if err != nil || lobby == nil {
		handler.sendTranslatedMessage(userID, message.Chat.ID, "error_lobby_not_found")
		return
	}

	argsParts := parseCommandArgs(args)
	if len(argsParts) == 0 {
		// List all payment methods
		methods, err := handler.paymentMethodService.GetPaymentMethodsByLobby(lobby.ID, false)
		if err != nil {
			handler.sendTranslatedMessage(userID, message.Chat.ID, "error_generic", err)
			return
		}

		if len(methods) == 0 {
			handler.sendTranslatedMessage(userID, message.Chat.ID, "payment_methods_none")
			return
		}

		var items []string
		for _, method := range methods {
			status := "✅"
			if !method.IsActive {
				status = "❌"
			}
			item := translator.T("payment_method_item", status, method.Name, method.Type)
			if method.ClosingDay.Valid {
				item += translator.T("payment_method_closing", method.ClosingDay.Int64)
			}
			if method.OwnerTelegramID.Valid {
				item += translator.T("payment_method_owner", method.OwnerTelegramID.Int64)
			}
			items = append(items, item)
		}
		msg := translator.T("payment_methods_list", strings.Join(items, "\n"))
		handler.sendMessage(message.Chat.ID, msg)
		return
	}

	action := strings.ToLower(argsParts[0])
	switch action {
	case "add":
		h.handleAddPaymentMethod(handler, message, lobby.ID, argsParts[1:])
	case "edit", "update":
		h.handleEditPaymentMethod(handler, message, argsParts[1:])
	case "delete", "remove":
		h.handleDeletePaymentMethod(handler, message, argsParts[1:])
	default:
		handler.sendTranslatedMessage(userID, message.Chat.ID, "payment_method_unknown_action")
	}
}

// handleAddPaymentMethod handles adding a payment method
func (h *Handler) handleAddPaymentMethod(handler *Handler, message *tgbotapi.Message, lobbyID int64, args []string) {
	if message.From == nil {
		handler.sendMessage(message.Chat.ID, "❌ Error: This command must be used by a user.")
		return
	}
	userID := message.From.ID
	translator := handler.getTranslator(userID)

	if len(args) < 2 {
		handler.sendTranslatedMessage(userID, message.Chat.ID, "payment_method_add_usage")
		return
	}

	name := args[0]
	methodType := strings.ToLower(args[1])
	var closingDay *int64
	var ownerID *int64

	// Parse closing day if provided
	if len(args) >= 3 {
		cd, err := strconv.ParseInt(args[2], 10, 64)
		if err != nil || cd < 1 || cd > 31 {
			handler.sendTranslatedMessage(userID, message.Chat.ID, "payment_method_closing_invalid")
			return
		}
		closingDay = &cd
	}

	// For credit cards, closing day is required
	if methodType == "credit_card" && closingDay == nil {
		handler.sendTranslatedMessage(userID, message.Chat.ID, "payment_method_closing_required")
		return
	}

	method, err := handler.paymentMethodService.CreatePaymentMethod(
		lobbyID, name, methodType, ownerID, closingDay)
	if err != nil {
		handler.sendTranslatedMessage(userID, message.Chat.ID, "payment_method_add_error", err)
		return
	}

	msg := translator.T("payment_method_added", method.Name)
	if method.ClosingDay.Valid {
		msg += translator.T("payment_method_closing_day", method.ClosingDay.Int64)
	}
	handler.sendMessage(message.Chat.ID, msg)
}

// handleEditPaymentMethod handles editing a payment method
func (h *Handler) handleEditPaymentMethod(handler *Handler, message *tgbotapi.Message, args []string) {
	if message.From == nil {
		handler.sendMessage(message.Chat.ID, "❌ Error: This command must be used by a user.")
		return
	}
	userID := message.From.ID

	if len(args) < 2 {
		handler.sendTranslatedMessage(userID, message.Chat.ID, "payment_method_edit_usage")
		return
	}

	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		handler.sendTranslatedMessage(userID, message.Chat.ID, "payment_method_invalid_id")
		return
	}

	field := strings.ToLower(args[1])
	var name *string
	var methodType *string
	var closingDay *int64
	var isActive *bool

	switch field {
	case "name":
		if len(args) < 3 {
			handler.sendTranslatedMessage(userID, message.Chat.ID, "payment_method_edit_usage")
			return
		}
		n := args[2]
		name = &n

	case "type":
		if len(args) < 3 {
			handler.sendTranslatedMessage(userID, message.Chat.ID, "payment_method_edit_usage")
			return
		}
		mt := strings.ToLower(args[2])
		methodType = &mt

	case "closing_day":
		if len(args) < 3 {
			handler.sendTranslatedMessage(userID, message.Chat.ID, "payment_method_edit_usage")
			return
		}
		cd, err := strconv.ParseInt(args[2], 10, 64)
		if err != nil || cd < 1 || cd > 31 {
			handler.sendTranslatedMessage(userID, message.Chat.ID, "payment_method_closing_invalid")
			return
		}
		closingDay = &cd

	case "active":
		if len(args) < 3 {
			handler.sendTranslatedMessage(userID, message.Chat.ID, "payment_method_edit_usage")
			return
		}
		active := strings.ToLower(args[2]) == "true"
		isActive = &active

	default:
		handler.sendTranslatedMessage(userID, message.Chat.ID, "payment_method_edit_usage")
		return
	}

	err = handler.paymentMethodService.UpdatePaymentMethod(id, name, methodType, nil, closingDay, isActive)
	if err != nil {
		handler.sendTranslatedMessage(userID, message.Chat.ID, "payment_method_update_error", err)
		return
	}

	handler.sendTranslatedMessage(userID, message.Chat.ID, "payment_method_updated")
}

// handleDeletePaymentMethod handles deleting a payment method
func (h *Handler) handleDeletePaymentMethod(handler *Handler, message *tgbotapi.Message, args []string) {
	if message.From == nil {
		handler.sendMessage(message.Chat.ID, "❌ Error: This command must be used by a user.")
		return
	}
	userID := message.From.ID

	if len(args) < 1 {
		handler.sendTranslatedMessage(userID, message.Chat.ID, "payment_method_delete_usage")
		return
	}

	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		handler.sendTranslatedMessage(userID, message.Chat.ID, "payment_method_invalid_id")
		return
	}

	err = handler.paymentMethodService.DeletePaymentMethod(id)
	if err != nil {
		handler.sendTranslatedMessage(userID, message.Chat.ID, "payment_method_delete_error", err)
		return
	}

	handler.sendTranslatedMessage(userID, message.Chat.ID, "payment_method_deleted")
}
