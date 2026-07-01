package bot

import (
	"botGastosPareja/pkg/utils"
	"fmt"
	"log"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// registerExportCommands registers export-related commands
func (h *Handler) registerExportCommands() {
	h.router.RegisterCommand("export_ynab", h.handleExportYNAB)
}

// handleExportYNAB handles the /export_ynab command — exports expenses as YNAB-compatible TSV
func (h *Handler) handleExportYNAB(handler *Handler, message *tgbotapi.Message, args string) {
	userID := message.From.ID
	translator := handler.getTranslator(userID)

	// Must be in a lobby
	lobby, err := handler.getLobbyForMessage(message)
	if err != nil || lobby == nil {
		handler.sendTranslatedMessage(userID, message.Chat.ID, "error_lobby_not_found")
		return
	}

	// Parse optional month argument
	var startDate, endDate time.Time
	argsParts := parseCommandArgs(args)

	if len(argsParts) >= 1 {
		// Try to parse as month
		monthTime, err := utils.ParseMonth(argsParts[0])
		if err == nil {
			start, end := utils.GetMonthStartEnd(monthTime.Year(), monthTime.Month())
			startDate = start
			endDate = end
		} else {
			// Try as date range: start end
			if len(argsParts) >= 2 {
				start, err1 := utils.ParseDate(argsParts[0])
				end, err2 := utils.ParseDate(argsParts[1])
				if err1 == nil && err2 == nil {
					startDate = start
					endDate = end
				}
			}
		}
	}

	// Default to current month if no valid period
	if startDate.IsZero() {
		now := time.Now()
		start, end := utils.GetMonthStartEnd(now.Year(), now.Month())
		startDate = start
		endDate = end
	}

	// Fetch expenses
	expenses, err := handler.expenseService.GetExpensesByLobby(lobby.ID, &startDate, &endDate, nil)
	if err != nil {
		log.Printf("Error fetching expenses for export: %v", err)
		handler.sendTranslatedMessage(userID, message.Chat.ID, "error_generic", err)
		return
	}

	if len(expenses) == 0 {
		// No expenses — send a text message instead of an empty file
		var msg string
		if translator.GetLanguage() == "es_AR" {
			msg = fmt.Sprintf("📭 No hay gastos en %s.", utils.FormatMonth(startDate))
		} else {
			msg = fmt.Sprintf("📭 No expenses in %s.", utils.FormatMonth(startDate))
		}
		handler.sendMessage(message.Chat.ID, msg)
		return
	}

	// Build TSV content — YNAB-compatible format
	// Header: Date, Payee, Memo, Amount
	// Amount is negative for outflows (expenses)
	var b strings.Builder
	b.WriteString("Date\tPayee\tMemo\tAmount\n")

	for _, exp := range expenses {
		date := exp.ExpenseDate.Format("2006-01-02")

		payee := ""
		if exp.Description.Valid {
			payee = cleanTSVField(exp.Description.String)
		}

		memo := ""
		if exp.Category.Valid {
			memo = cleanTSVField(exp.Category.String)
		}

		// Amount is always a negative outflow (expense)
		amount := fmt.Sprintf("-%.2f", exp.Amount)

		b.WriteString(fmt.Sprintf("%s\t%s\t%s\t%s\n", date, payee, memo, amount))
	}

	tsvContent := []byte(b.String())

	// Send as file attachment
	fileName := fmt.Sprintf("saldaando_export_%s.tsv", startDate.Format("2006-01"))
	if !startDate.IsZero() && !endDate.IsZero() {
		if startDate.Month() == endDate.Month() && startDate.Year() == endDate.Year() {
			fileName = fmt.Sprintf("saldaando_export_%s.tsv", startDate.Format("2006-01"))
		} else {
			fileName = fmt.Sprintf("saldaando_export_%s_%s.tsv",
				startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
		}
	}

	doc := tgbotapi.NewDocument(message.Chat.ID, tgbotapi.FileBytes{
		Name:  fileName,
		Bytes: tsvContent,
	})

	if _, err := handler.bot.Send(doc); err != nil {
		log.Printf("Error sending TSV document: %v", err)
		handler.sendTranslatedMessage(userID, message.Chat.ID, "error_generic", "Failed to send export file")
		return
	}

	log.Printf("Exported %d expenses to TSV for lobby %d, period %s to %s",
		len(expenses), lobby.ID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
}

// cleanTSVField prepares a string for TSV output
func cleanTSVField(s string) string {
	// Replace tabs and newlines to avoid breaking the TSV structure
	s = strings.ReplaceAll(s, "\t", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", "")
	return s
}
