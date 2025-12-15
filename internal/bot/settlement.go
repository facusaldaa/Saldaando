package bot

import (
	"botGastosPareja/internal/database"
	"botGastosPareja/internal/service"
	"botGastosPareja/pkg/i18n"
	"botGastosPareja/pkg/utils"
	"fmt"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// registerSettlementCommands registers settlement-related commands
func (h *Handler) registerSettlementCommands() {
	h.router.RegisterCommand("settle", h.handleSettle)
	h.router.RegisterCommand("settle_billing", h.handleSettleBilling)
}

// handleSettle handles the /settle command
func (h *Handler) handleSettle(handler *Handler, message *tgbotapi.Message, args string) {
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

	if len(argsParts) >= 1 {
		// Parse start date
		start, err := utils.ParseMonth(argsParts[0])
		if err == nil {
			startTime, endTime := utils.GetMonthStartEnd(start.Year(), start.Month())
			startDate = &startTime
			endDate = &endTime
		} else {
			// Try parsing as date range
			if len(argsParts) >= 2 {
				start, err1 := utils.ParseDate(argsParts[0])
				end, err2 := utils.ParseDate(argsParts[1])
				if err1 == nil && err2 == nil {
					startDate = &start
					endDate = &end
				}
			}
		}
	}

	if startDate == nil {
		// Default to current month
		now := time.Now()
		start, end := utils.GetMonthStartEnd(now.Year(), now.Month())
		startDate = &start
		endDate = &end
	}

	result, err := handler.settlementService.CalculateSettlement(lobby.ID, startDate, endDate)
	if err != nil {
		handler.sendTranslatedMessage(userID, message.Chat.ID, "settle_error", err)
		return
	}

	msg := h.formatSettlementResult(result, startDate, endDate, translator)
	handler.sendMessage(message.Chat.ID, msg)
}

// handleSettleBilling handles the /settle_billing command
func (h *Handler) handleSettleBilling(handler *Handler, message *tgbotapi.Message, args string) {
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
		handler.sendTranslatedMessage(userID, message.Chat.ID, "settle_usage")
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

	result, err := handler.settlementService.CalculateBillingSettlement(
		lobby.ID, paymentMethod.ID, periodStart, periodEnd)
	if err != nil {
		handler.sendTranslatedMessage(userID, message.Chat.ID, "settle_error", err)
		return
	}

	msg := h.formatSettlementResult(result, &periodStart, &periodEnd, translator)
	handler.sendMessage(message.Chat.ID, msg)
}

// formatSettlementResult formats a settlement result for display
func (h *Handler) formatSettlementResult(result *service.SettlementResult, startDate, endDate *time.Time, translator *i18n.Translator) string {
	periodStr := translator.T("summary_period")
	if startDate != nil && endDate != nil {
		periodStr = fmt.Sprintf("%s to %s",
			utils.FormatDate(*startDate), utils.FormatDate(*endDate))
	}

	msg := translator.T("settle_report",
		periodStr,
		result.AccountType,
		utils.FormatCurrency(result.TotalExpenses))

	if result.AccountType == "separate" {
		msg += translator.T("settle_separate")
	} else {
		msg += translator.T("settle_shared")
	}

	msg += translator.T("settle_user1_spent", utils.FormatCurrency(result.User1TotalSpent))
	msg += translator.T("settle_user2_spent", utils.FormatCurrency(result.User2TotalSpent))

	if result.AccountType == "shared" {
		// Use actual salary percentages from lobby, not calculated from expected
		msg += translator.T("settle_user1_expected",
			result.User1SalaryPercentage*100, utils.FormatCurrency(result.User1Expected))
		msg += translator.T("settle_user2_expected",
			result.User2SalaryPercentage*100, utils.FormatCurrency(result.User2Expected))
	} else {
		expectedPerPerson := result.TotalExpenses / 2.0
		msg += translator.T("settle_expected_per", utils.FormatCurrency(expectedPerPerson))
	}

	// Determine who owes whom
	if result.User1Debt > 0 {
		msg += translator.T("settle_user1_owes", utils.FormatCurrency(result.User1Debt))
	} else if result.User1Debt < 0 {
		msg += translator.T("settle_user2_owes", utils.FormatCurrency(-result.User1Debt))
	} else {
		msg += translator.T("settle_all_settled")
	}

	if result.User2Debt > 0 {
		msg += translator.T("settle_user2_owes", utils.FormatCurrency(result.User2Debt))
	} else if result.User2Debt < 0 {
		msg += translator.T("settle_user1_owes", utils.FormatCurrency(-result.User2Debt))
	}

	return msg
}
