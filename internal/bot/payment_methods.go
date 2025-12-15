package bot

import (
	"fmt"
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
	// Get user's lobby for this specific chat (group/private)
	lobby, err := handler.getLobbyForMessage(message)
	if err != nil || lobby == nil {
		handler.sendMessage(message.Chat.ID,
			"❌ You're not in a lobby yet. Use /start to create or join one.")
		return
	}

	argsParts := parseCommandArgs(args)
	if len(argsParts) == 0 {
		// List all payment methods
		methods, err := handler.paymentMethodService.GetPaymentMethodsByLobby(lobby.ID, false)
		if err != nil {
			handler.sendMessage(message.Chat.ID,
				fmt.Sprintf("❌ Error: %v", err))
			return
		}

		if len(methods) == 0 {
			handler.sendMessage(message.Chat.ID,
				"📋 No payment methods configured.\n\n"+
					"Add one with:\n"+
					"`/payment_methods add <name> <type> [closing_day]`\n\n"+
					"Types: credit_card, debit_card, cash, bank_transfer, other")
			return
		}

		msg := "📋 *Payment Methods:*\n\n"
		for _, method := range methods {
			status := "✅"
			if !method.IsActive {
				status = "❌"
			}
			msg += fmt.Sprintf("%s *%s* (%s)", status, method.Name, method.Type)
			if method.ClosingDay.Valid {
				msg += fmt.Sprintf(" - Closes on %d", method.ClosingDay.Int64)
			}
			if method.OwnerTelegramID.Valid {
				msg += fmt.Sprintf(" - Owner: %d", method.OwnerTelegramID.Int64)
			}
			msg += "\n"
		}
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
		handler.sendMessage(message.Chat.ID,
			"❌ Unknown action. Use: `add`, `edit`, or `delete`")
	}
}

// handleAddPaymentMethod handles adding a payment method
func (h *Handler) handleAddPaymentMethod(handler *Handler, message *tgbotapi.Message, lobbyID int64, args []string) {
	if len(args) < 2 {
		handler.sendMessage(message.Chat.ID,
			"❌ Usage: `/payment_methods add <name> <type> [closing_day]`\n\n"+
				"Types: credit_card, debit_card, cash, bank_transfer, other\n"+
				"Example: `/payment_methods add Visa credit_card 15`")
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
			handler.sendMessage(message.Chat.ID,
				"❌ Closing day must be a number between 1 and 31")
			return
		}
		closingDay = &cd
	}

	// For credit cards, closing day is required
	if methodType == "credit_card" && closingDay == nil {
		handler.sendMessage(message.Chat.ID,
			"❌ Credit cards require a closing day. Usage: `/payment_methods add <name> credit_card <closing_day>`")
		return
	}

	method, err := handler.paymentMethodService.CreatePaymentMethod(
		lobbyID, name, methodType, ownerID, closingDay)
	if err != nil {
		handler.sendMessage(message.Chat.ID,
			fmt.Sprintf("❌ Failed to create payment method: %v", err))
		return
	}

	msg := fmt.Sprintf("✅ Payment method *%s* created successfully!", method.Name)
	if method.ClosingDay.Valid {
		msg += fmt.Sprintf("\nClosing day: %d", method.ClosingDay.Int64)
	}
	handler.sendMessage(message.Chat.ID, msg)
}

// handleEditPaymentMethod handles editing a payment method
func (h *Handler) handleEditPaymentMethod(handler *Handler, message *tgbotapi.Message, args []string) {
	if len(args) < 2 {
		handler.sendMessage(message.Chat.ID,
			"❌ Usage: `/payment_methods edit <id> <field> <value>`\n\n"+
				"Fields: name, type, closing_day, active\n"+
				"Example: `/payment_methods edit 1 closing_day 20`")
		return
	}

	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		handler.sendMessage(message.Chat.ID, "❌ Invalid payment method ID")
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
			handler.sendMessage(message.Chat.ID, "❌ Please provide a new name")
			return
		}
		n := args[2]
		name = &n

	case "type":
		if len(args) < 3 {
			handler.sendMessage(message.Chat.ID, "❌ Please provide a type")
			return
		}
		mt := strings.ToLower(args[2])
		methodType = &mt

	case "closing_day":
		if len(args) < 3 {
			handler.sendMessage(message.Chat.ID, "❌ Please provide a closing day (1-31)")
			return
		}
		cd, err := strconv.ParseInt(args[2], 10, 64)
		if err != nil || cd < 1 || cd > 31 {
			handler.sendMessage(message.Chat.ID, "❌ Closing day must be between 1 and 31")
			return
		}
		closingDay = &cd

	case "active":
		if len(args) < 3 {
			handler.sendMessage(message.Chat.ID, "❌ Please provide true or false")
			return
		}
		active := strings.ToLower(args[2]) == "true"
		isActive = &active

	default:
		handler.sendMessage(message.Chat.ID,
			"❌ Unknown field. Use: name, type, closing_day, or active")
		return
	}

	err = handler.paymentMethodService.UpdatePaymentMethod(id, name, methodType, nil, closingDay, isActive)
	if err != nil {
		handler.sendMessage(message.Chat.ID,
			fmt.Sprintf("❌ Failed to update payment method: %v", err))
		return
	}

	handler.sendMessage(message.Chat.ID, "✅ Payment method updated successfully!")
}

// handleDeletePaymentMethod handles deleting a payment method
func (h *Handler) handleDeletePaymentMethod(handler *Handler, message *tgbotapi.Message, args []string) {
	if len(args) < 1 {
		handler.sendMessage(message.Chat.ID,
			"❌ Usage: `/payment_methods delete <id>`")
		return
	}

	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		handler.sendMessage(message.Chat.ID, "❌ Invalid payment method ID")
		return
	}

	err = handler.paymentMethodService.DeletePaymentMethod(id)
	if err != nil {
		handler.sendMessage(message.Chat.ID,
			fmt.Sprintf("❌ Failed to delete payment method: %v", err))
		return
	}

	handler.sendMessage(message.Chat.ID, "✅ Payment method deleted successfully!")
}
