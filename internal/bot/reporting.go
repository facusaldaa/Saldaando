package bot

import (
	"botGastosPareja/internal/database"
	"botGastosPareja/pkg/i18n"
	"botGastosPareja/pkg/utils"
	"fmt"
	"sort"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// registerReportingCommands registers reporting-related commands
func (h *Handler) registerReportingCommands() {
	h.router.RegisterCommand("summary", h.handleSummary)
	h.router.RegisterCommand("summary_billing", h.handleSummaryBilling)
}

// handleSummary handles the /summary command
func (h *Handler) handleSummary(handler *Handler, message *tgbotapi.Message, args string) {
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

	if len(argsParts) >= 2 {
		// Parse date range
		start, err1 := utils.ParseDate(argsParts[0])
		end, err2 := utils.ParseDate(argsParts[1])
		if err1 == nil && err2 == nil {
			startDate = &start
			endDate = &end
		}
	} else if len(argsParts) >= 1 {
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

	msg := h.formatSummary(expenses, lobby, startDate, endDate, translator)
	handler.sendMessage(message.Chat.ID, msg)
}

// handleSummaryBilling handles the /summary_billing command
func (h *Handler) handleSummaryBilling(handler *Handler, message *tgbotapi.Message, args string) {
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
		handler.sendTranslatedMessage(userID, message.Chat.ID, "summary_billing_usage")
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

	msg := h.formatSummary(expenses, lobby, &periodStart, &periodEnd, translator)
	handler.sendMessage(message.Chat.ID, msg)
}

// formatSummary formats a summary report
func (h *Handler) formatSummary(expenses []*database.Expense, lobby *database.Lobby, startDate, endDate *time.Time, translator *i18n.Translator) string {
	if len(expenses) == 0 {
		periodStr := translator.T("summary_period")
		if startDate != nil && endDate != nil {
			periodStr = fmt.Sprintf("%s to %s",
				utils.FormatDate(*startDate), utils.FormatDate(*endDate))
		}
		return translator.T("summary_none", periodStr)
	}

	periodStr := translator.T("summary_period")
	if startDate != nil && endDate != nil {
		periodStr = fmt.Sprintf("%s to %s",
			utils.FormatDate(*startDate), utils.FormatDate(*endDate))
	}

	// Get user names
	user1Name := h.getUserDisplayName(lobby.User1TelegramID, "Usuario 1")
	user2Name := h.getUserDisplayName(lobby.User2TelegramID, "Usuario 2")

	var total float64
	user1Total := 0.0
	user2Total := 0.0
	categoryTotals := make(map[string]float64)
	paymentMethodTotals := make(map[string]float64)

	for _, exp := range expenses {
		total += exp.Amount

		if exp.SpenderTelegramID == lobby.User1TelegramID {
			user1Total += exp.Amount
		} else if exp.SpenderTelegramID == lobby.User2TelegramID {
			user2Total += exp.Amount
		}

		if exp.Category.Valid {
			categoryTotals[exp.Category.String] += exp.Amount
		}

		if exp.PaymentMethodID.Valid {
			pm, _ := h.paymentMethodService.GetPaymentMethodByID(exp.PaymentMethodID.Int64)
			if pm != nil {
				paymentMethodTotals[pm.Name] += exp.Amount
			}
		}
	}

	msg := translator.T("summary_header",
		periodStr,
		utils.FormatCurrency(total),
		len(expenses))

	// Per-person breakdown
	msg += translator.T("summary_by_person")
	msg += fmt.Sprintf("%s: %s (%.1f%%)\n", user1Name, utils.FormatCurrency(user1Total), user1Total/total*100)
	msg += fmt.Sprintf("%s: %s (%.1f%%)\n\n", user2Name, utils.FormatCurrency(user2Total), user2Total/total*100)

	// Category breakdown
	if len(categoryTotals) > 0 {
		msg += translator.T("summary_by_category")
		type catTotal struct {
			name  string
			total float64
		}
		cats := make([]catTotal, 0, len(categoryTotals))
		for name, total := range categoryTotals {
			cats = append(cats, catTotal{name, total})
		}
		sort.Slice(cats, func(i, j int) bool {
			return cats[i].total > cats[j].total
		})
		for _, cat := range cats {
			msg += translator.T("summary_category_item",
				cat.name, utils.FormatCurrency(cat.total), cat.total/total*100)
		}
		msg += "\n"
	}

	// Payment method breakdown
	if len(paymentMethodTotals) > 0 {
		msg += translator.T("summary_by_payment")
		type pmTotal struct {
			name  string
			total float64
		}
		pms := make([]pmTotal, 0, len(paymentMethodTotals))
		for name, total := range paymentMethodTotals {
			pms = append(pms, pmTotal{name, total})
		}
		sort.Slice(pms, func(i, j int) bool {
			return pms[i].total > pms[j].total
		})
		for _, pm := range pms {
			msg += fmt.Sprintf("• %s: %s (%.1f%%)\n",
				pm.name, utils.FormatCurrency(pm.total), pm.total/total*100)
		}
	}

	return msg
}
