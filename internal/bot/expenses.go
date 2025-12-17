package bot

import (
	"botGastosPareja/internal/database"
	"botGastosPareja/pkg/utils"
	"fmt"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// registerExpenseCommands registers expense-related commands
func (h *Handler) registerExpenseCommands() {
	h.router.RegisterCommand("add", h.handleAddExpense)
	h.router.RegisterCommand("list", h.handleListExpenses)
	h.router.RegisterCommand("list_billing", h.handleListBillingExpenses)
	h.router.RegisterCommand("delete", h.handleDeleteExpense)
	h.router.RegisterCommand("edit", h.handleEditExpense)
}

// handleAddExpense handles the /add command
func (h *Handler) handleAddExpense(handler *Handler, message *tgbotapi.Message, args string) {
	userID := message.From.ID
	translator := handler.getTranslator(userID)

	// Get user's lobby for this specific chat (group/private)
	lobby, err := handler.getLobbyForMessage(message)
	if err != nil || lobby == nil {
		handler.sendTranslatedMessage(userID, message.Chat.ID, "error_lobby_not_found")
		return
	}

	argsParts := parseCommandArgs(args)
	if len(argsParts) < 2 {
		handler.sendTranslatedMessage(userID, message.Chat.ID, "expense_add_usage")
		return
	}

	// Parse amount
	amount, err := strconv.ParseFloat(argsParts[0], 64)
	if err != nil || amount <= 0 {
		handler.sendTranslatedMessage(userID, message.Chat.ID, "expense_invalid_amount")
		return
	}

	description := argsParts[1]
	category := ""
	paymentMethodName := ""
	spenderArg := ""

	// Parse optional arguments in order: [category] [payment_method] [spender]
	// Spender is always the last optional argument
	argIndex := 2
	if len(argsParts) > argIndex {
		// Check if last argument is a spender identifier
		lastArg := strings.ToLower(argsParts[len(argsParts)-1])
		if lastArg == "user1" || lastArg == "user2" || lastArg == "partner" || lastArg == "pareja" {
			spenderArg = argsParts[len(argsParts)-1]
			// Remove spender from argsParts for category/payment method parsing
			argsParts = argsParts[:len(argsParts)-1]
		} else if _, err := strconv.ParseInt(argsParts[len(argsParts)-1], 10, 64); err == nil && len(argsParts[len(argsParts)-1]) > 3 {
			// If last arg is a long number (likely Telegram ID), treat as spender
			spenderArg = argsParts[len(argsParts)-1]
			argsParts = argsParts[:len(argsParts)-1]
		}
	}

	// Parse category and payment method from remaining args
	if len(argsParts) > argIndex {
		category = argsParts[argIndex]
		argIndex++
	}
	if len(argsParts) > argIndex {
		paymentMethodName = argsParts[argIndex]
	}

	// Determine spender ID
	spenderID := userID // Default to the user adding the expense
	if spenderArg != "" {
		spenderArgLower := strings.ToLower(spenderArg)
		if spenderArgLower == "user2" || spenderArgLower == "partner" || spenderArgLower == "pareja" {
			// Add expense for the other user (user2)
			if lobby.User2TelegramID != 0 {
				spenderID = lobby.User2TelegramID
			} else {
				handler.sendTranslatedMessage(userID, message.Chat.ID, "waiting_partner")
				return
			}
		} else if spenderArgLower == "user1" {
			// Explicitly add for user1
			spenderID = lobby.User1TelegramID
		} else if parsedID, err := strconv.ParseInt(spenderArg, 10, 64); err == nil {
			// Check if it's a valid user ID in the lobby
			if parsedID == lobby.User1TelegramID || parsedID == lobby.User2TelegramID {
				spenderID = parsedID
			} else {
				handler.sendTranslatedMessage(userID, message.Chat.ID, "error_invalid_user_id")
				return
			}
		}
	}

	var paymentMethodID *int64
	if paymentMethodName != "" {
		// Find payment method by name
		methods, err := handler.paymentMethodService.GetPaymentMethodsByLobby(lobby.ID, true)
		if err == nil {
			for _, method := range methods {
				if strings.EqualFold(method.Name, paymentMethodName) {
					paymentMethodID = &method.ID
					break
				}
			}
		}
		if paymentMethodID == nil {
			// Show available payment methods
			if len(methods) > 0 {
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
					items = append(items, item)
				}
				handler.sendTranslatedMessage(userID, message.Chat.ID, "payment_method_not_found_list", paymentMethodName, strings.Join(items, "\n"))
			} else {
				handler.sendTranslatedMessage(userID, message.Chat.ID, "payment_method_not_found", paymentMethodName)
			}
		}
	}

	expenseDate := time.Now()
	expense, err := handler.expenseService.CreateExpense(
		lobby.ID,
		spenderID, // Use the determined spender ID
		amount,
		description,
		category,
		expenseDate,
		paymentMethodID,
	)
	if err != nil {
		handler.sendTranslatedMessage(userID, message.Chat.ID, "expense_add_error", err)
		return
	}

	msg := translator.T("expense_added",
		utils.FormatCurrency(expense.Amount),
		expense.Description.String)
	msg += fmt.Sprintf("ID: %d\n", expense.ID)

	if expense.Category.Valid {
		msg += translator.T("expense_category", expense.Category.String)
	}
	if expense.PaymentMethodID.Valid {
		pm, _ := handler.paymentMethodService.GetPaymentMethodByID(expense.PaymentMethodID.Int64)
		if pm != nil {
			msg += translator.T("expense_payment_method", pm.Name)
		}
	}
	if expense.BillingPeriodStart.Valid {
		msg += translator.T("expense_billing_period",
			utils.FormatDate(expense.BillingPeriodStart.Time),
			utils.FormatDate(expense.BillingPeriodEnd.Time))
	}

	handler.sendMessage(message.Chat.ID, msg)
}

// handleListExpenses handles the /list command
func (h *Handler) handleListExpenses(handler *Handler, message *tgbotapi.Message, args string) {
	userID := message.From.ID
	translator := handler.getTranslator(userID)

	// Get user's lobby for this specific chat (group/private)
	lobby, err := handler.getLobbyForMessage(message)
	if err != nil || lobby == nil {
		handler.sendTranslatedMessage(userID, message.Chat.ID, "error_lobby_not_found")
		return
	}

	var startDate, endDate *time.Time
	argsParts := parseCommandArgs(args)

	if len(argsParts) > 0 {
		// Parse month
		monthTime, err := utils.ParseMonth(argsParts[0])
		if err == nil {
			start, end := utils.GetMonthStartEnd(monthTime.Year(), monthTime.Month())
			startDate = &start
			endDate = &end
		}
	} else {
		// Default to current month
		now := time.Now()
		start, end := utils.GetMonthStartEnd(now.Year(), now.Month())
		startDate = &start
		endDate = &end
	}

	expenses, err := handler.expenseService.GetExpensesByLobby(lobby.ID, startDate, endDate, nil)
	if err != nil {
		handler.sendTranslatedMessage(userID, message.Chat.ID, "error_generic", err)
		return
	}

	if len(expenses) == 0 {
		period := translator.T("summary_period")
		if startDate != nil {
			period = utils.FormatMonth(*startDate)
		}
		handler.sendTranslatedMessage(userID, message.Chat.ID, "expense_list_none", period)
		return
	}

	// Get user names for display
	user1, _ := handler.userService.GetUserByTelegramID(lobby.User1TelegramID)
	user2, _ := handler.userService.GetUserByTelegramID(lobby.User2TelegramID)

	var total float64
	msg := translator.T("expense_list_header", len(expenses))
	for _, exp := range expenses {
		total += exp.Amount
		desc := exp.Description.String
		if !exp.Description.Valid {
			desc = translator.T("expense_no_description")
		}

		// Get the spender's name
		var userLabel string
		if exp.SpenderTelegramID == lobby.User1TelegramID {
			if user1 != nil && user1.DisplayName.Valid && user1.DisplayName.String != "" {
				userLabel = user1.DisplayName.String
			} else if user1 != nil && user1.Username.Valid && user1.Username.String != "" {
				userLabel = "@" + user1.Username.String
			} else {
				userLabel = "User 1"
			}
		} else if exp.SpenderTelegramID == lobby.User2TelegramID {
			if user2 != nil && user2.DisplayName.Valid && user2.DisplayName.String != "" {
				userLabel = user2.DisplayName.String
			} else if user2 != nil && user2.Username.Valid && user2.Username.String != "" {
				userLabel = "@" + user2.Username.String
			} else {
				userLabel = "User 2"
			}
		} else {
			// Try to get the user
			spender, _ := handler.userService.GetUserByTelegramID(exp.SpenderTelegramID)
			if spender != nil && spender.DisplayName.Valid && spender.DisplayName.String != "" {
				userLabel = spender.DisplayName.String
			} else if spender != nil && spender.Username.Valid && spender.Username.String != "" {
				userLabel = "@" + spender.Username.String
			} else {
				userLabel = fmt.Sprintf("User %d", exp.SpenderTelegramID)
			}
		}

		msg += fmt.Sprintf("[ID: %d] ", exp.ID)
		msg += translator.T("expense_list_item", utils.FormatCurrency(exp.Amount), desc)
		msg += fmt.Sprintf("  Added by: %s\n", userLabel)
		if exp.Category.Valid {
			msg += translator.T("expense_list_category", exp.Category.String)
		}
		if exp.PaymentMethodID.Valid {
			pm, _ := handler.paymentMethodService.GetPaymentMethodByID(exp.PaymentMethodID.Int64)
			if pm != nil {
				msg += translator.T("expense_payment_method", pm.Name)
			}
		}
		msg += translator.T("expense_list_date", utils.FormatDate(exp.ExpenseDate))
	}

	msg += translator.T("expense_list_total", utils.FormatCurrency(total))
	handler.sendMessage(message.Chat.ID, msg)
}

// handleListBillingExpenses handles the /list_billing command
func (h *Handler) handleListBillingExpenses(handler *Handler, message *tgbotapi.Message, args string) {
	userID := message.From.ID
	translator := handler.getTranslator(userID)

	// Get user's lobby for this specific chat (group/private)
	lobby, err := handler.getLobbyForMessage(message)
	if err != nil || lobby == nil {
		handler.sendTranslatedMessage(userID, message.Chat.ID, "error_lobby_not_found")
		return
	}

	argsParts := parseCommandArgs(args)
	if len(argsParts) < 1 {
		handler.sendTranslatedMessage(userID, message.Chat.ID, "expense_billing_usage")
		return
	}

	paymentMethodName := argsParts[0]
	methods, err := handler.paymentMethodService.GetPaymentMethodsByLobby(lobby.ID, true)
	if err != nil {
		handler.sendTranslatedMessage(userID, message.Chat.ID, "error_generic", err)
		return
	}

	var paymentMethod *database.PaymentMethod
	for _, method := range methods {
		if strings.EqualFold(method.Name, paymentMethodName) {
			paymentMethod = method
			break
		}
	}

	if paymentMethod == nil {
		handler.sendTranslatedMessage(userID, message.Chat.ID, "payment_method_not_found", paymentMethodName)
		return
	}

	if !paymentMethod.ClosingDay.Valid {
		handler.sendTranslatedMessage(userID, message.Chat.ID, "expense_billing_no_cycle")
		return
	}

	// Parse period or use current
	var periodStart, periodEnd time.Time
	if len(argsParts) >= 2 {
		monthTime, err := utils.ParseMonth(argsParts[1])
		if err != nil {
			handler.sendTranslatedMessage(userID, message.Chat.ID, "error_invalid_period")
			return
		}
		periodStart, periodEnd = utils.GetBillingPeriodForMonth(
			monthTime.Year(), monthTime.Month(), int(paymentMethod.ClosingDay.Int64))
	} else {
		now := time.Now()
		periodStart, periodEnd = utils.GetBillingPeriodForMonth(
			now.Year(), now.Month(), int(paymentMethod.ClosingDay.Int64))
	}

	expenses, err := handler.expenseService.GetExpensesByBillingPeriod(
		lobby.ID, paymentMethod.ID, periodStart, periodEnd)
	if err != nil {
		handler.sendTranslatedMessage(userID, message.Chat.ID, "error_generic", err)
		return
	}

	if len(expenses) == 0 {
		handler.sendTranslatedMessage(userID, message.Chat.ID, "expense_billing_none",
			utils.FormatDate(periodStart), utils.FormatDate(periodEnd))
		return
	}

	var total float64
	msg := translator.T("expense_billing_header",
		paymentMethod.Name,
		utils.FormatDate(periodStart),
		utils.FormatDate(periodEnd))

	for _, exp := range expenses {
		total += exp.Amount
		desc := exp.Description.String
		if !exp.Description.Valid {
			desc = translator.T("expense_no_description")
		}
		msg += fmt.Sprintf("[ID: %d] • %s - %s (%s)\n",
			exp.ID,
			utils.FormatCurrency(exp.Amount),
			desc,
			utils.FormatDate(exp.ExpenseDate))
	}

	msg += fmt.Sprintf("\n*Total: %s*", utils.FormatCurrency(total))
	handler.sendMessage(message.Chat.ID, msg)
}

// handleDeleteExpense handles the /delete command
func (h *Handler) handleDeleteExpense(handler *Handler, message *tgbotapi.Message, args string) {
	userID := message.From.ID
	translator := handler.getTranslator(userID)

	// Get user's lobby for this specific chat (group/private)
	lobby, err := handler.getLobbyForMessage(message)
	if err != nil || lobby == nil {
		handler.sendTranslatedMessage(userID, message.Chat.ID, "error_lobby_not_found")
		return
	}

	argsParts := parseCommandArgs(args)
	if len(argsParts) < 1 {
		// Show recent expenses for selection
		now := time.Now()
		start, end := utils.GetMonthStartEnd(now.Year(), now.Month())
		expenses, err := handler.expenseService.GetExpensesByLobby(lobby.ID, &start, &end, nil)
		if err != nil {
			handler.sendTranslatedMessage(userID, message.Chat.ID, "error_generic", err)
			return
		}

		if len(expenses) == 0 {
			handler.sendTranslatedMessage(userID, message.Chat.ID, "expense_delete_none")
			return
		}

		// Show last 10 expenses
		maxShow := 10
		if len(expenses) < maxShow {
			maxShow = len(expenses)
		}

		msg := translator.T("expense_delete_list_header")
		for i := 0; i < maxShow; i++ {
			exp := expenses[i]
			desc := exp.Description.String
			if !exp.Description.Valid {
				desc = translator.T("expense_no_description")
			}
			cat := ""
			if exp.Category.Valid {
				cat = " | " + exp.Category.String
			}
			pm := ""
			if exp.PaymentMethodID.Valid {
				pmObj, _ := handler.paymentMethodService.GetPaymentMethodByID(exp.PaymentMethodID.Int64)
				if pmObj != nil {
					pm = " | " + pmObj.Name
				}
			}
			msg += fmt.Sprintf("%d. %s - %s%s%s (%s)\n",
				exp.ID,
				utils.FormatCurrency(exp.Amount),
				desc,
				cat,
				pm,
				utils.FormatDate(exp.ExpenseDate))
		}
		msg += translator.T("expense_delete_usage")
		handler.sendMessage(message.Chat.ID, msg)
		return
	}

	// Parse expense ID
	expenseID, err := strconv.ParseInt(argsParts[0], 10, 64)
	if err != nil || expenseID <= 0 {
		handler.sendTranslatedMessage(userID, message.Chat.ID, "expense_delete_invalid_id")
		return
	}

	// Verify expense belongs to lobby
	expense, err := handler.expenseService.GetExpenseByID(expenseID)
	if err != nil || expense == nil {
		handler.sendTranslatedMessage(userID, message.Chat.ID, "expense_delete_not_found")
		return
	}

	if expense.LobbyID != lobby.ID {
		handler.sendTranslatedMessage(userID, message.Chat.ID, "expense_delete_not_found")
		return
	}

	// Delete the expense
	err = handler.expenseService.DeleteExpense(expenseID)
	if err != nil {
		handler.sendTranslatedMessage(userID, message.Chat.ID, "expense_delete_error", err)
		return
	}

	handler.sendTranslatedMessage(userID, message.Chat.ID, "expense_deleted")
}

// handleEditExpense handles the /edit command
func (h *Handler) handleEditExpense(handler *Handler, message *tgbotapi.Message, args string) {
	userID := message.From.ID
	translator := handler.getTranslator(userID)

	// Get user's lobby for this specific chat (group/private)
	lobby, err := handler.getLobbyForMessage(message)
	if err != nil || lobby == nil {
		handler.sendTranslatedMessage(userID, message.Chat.ID, "error_lobby_not_found")
		return
	}

	argsParts := parseCommandArgs(args)
	if len(argsParts) < 2 {
		handler.sendTranslatedMessage(userID, message.Chat.ID, "expense_edit_usage")
		return
	}

	// Parse expense ID
	expenseID, err := strconv.ParseInt(argsParts[0], 10, 64)
	if err != nil || expenseID <= 0 {
		handler.sendTranslatedMessage(userID, message.Chat.ID, "expense_edit_invalid_id")
		return
	}

	// Verify expense belongs to lobby
	expense, err := handler.expenseService.GetExpenseByID(expenseID)
	if err != nil || expense == nil {
		handler.sendTranslatedMessage(userID, message.Chat.ID, "expense_edit_not_found")
		return
	}

	if expense.LobbyID != lobby.ID {
		handler.sendTranslatedMessage(userID, message.Chat.ID, "expense_edit_not_found")
		return
	}

	// Parse what to edit: field value
	field := strings.ToLower(argsParts[1])
	var category *string
	var paymentMethodID *int64

	if field == "category" {
		if len(argsParts) < 3 {
			handler.sendTranslatedMessage(userID, message.Chat.ID, "expense_edit_category_usage")
			return
		}
		cat := strings.Join(argsParts[2:], " ")
		// Allow empty string to clear category
		category = &cat
	} else if field == "payment_method" || field == "payment" {
		if len(argsParts) < 3 {
			handler.sendTranslatedMessage(userID, message.Chat.ID, "expense_edit_payment_usage")
			return
		}
		paymentMethodName := argsParts[2]
		// Find payment method by name
		methods, err := handler.paymentMethodService.GetPaymentMethodsByLobby(lobby.ID, true)
		if err == nil {
			for _, method := range methods {
				if strings.EqualFold(method.Name, paymentMethodName) {
					paymentMethodID = &method.ID
					break
				}
			}
		}
		if paymentMethodID == nil {
			handler.sendTranslatedMessage(userID, message.Chat.ID, "payment_method_not_found", paymentMethodName)
			return
		}
	} else {
		handler.sendTranslatedMessage(userID, message.Chat.ID, "expense_edit_invalid_field")
		return
	}

	// Update the expense
	err = handler.expenseService.UpdateExpense(expenseID, nil, nil, category, nil, paymentMethodID)
	if err != nil {
		handler.sendTranslatedMessage(userID, message.Chat.ID, "expense_edit_error", err)
		return
	}

	// Get updated expense to show confirmation
	updatedExpense, _ := handler.expenseService.GetExpenseByID(expenseID)
	msg := translator.T("expense_edited")
	if updatedExpense != nil {
		msg += fmt.Sprintf("\n\nID: %d\nAmount: %s\nDescription: %s\n",
			updatedExpense.ID,
			utils.FormatCurrency(updatedExpense.Amount),
			updatedExpense.Description.String)
		if updatedExpense.Category.Valid {
			msg += translator.T("expense_category", updatedExpense.Category.String)
		}
		if updatedExpense.PaymentMethodID.Valid {
			pm, _ := handler.paymentMethodService.GetPaymentMethodByID(updatedExpense.PaymentMethodID.Int64)
			if pm != nil {
				msg += translator.T("expense_payment_method", pm.Name)
			}
		}
	}

	handler.sendMessage(message.Chat.ID, msg)
}
