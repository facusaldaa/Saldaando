package bot

import (
	"botGastosPareja/internal/database"
	"botGastosPareja/internal/service"
	"botGastosPareja/pkg/i18n"
	"botGastosPareja/pkg/utils"
	"fmt"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// SettlementScheduler handles automatic monthly settlements
type SettlementScheduler struct {
	bot               *tgbotapi.BotAPI
	lobbyService      *service.LobbyService
	settlementService *service.SettlementService
	expenseService    *service.ExpenseService
	userService       *service.UserService
	lastCheckDate     time.Time
}

// NewSettlementScheduler creates a new settlement scheduler
func NewSettlementScheduler(bot *tgbotapi.BotAPI, lobbyService *service.LobbyService, settlementService *service.SettlementService, expenseService *service.ExpenseService, userService *service.UserService) *SettlementScheduler {
	return &SettlementScheduler{
		bot:               bot,
		lobbyService:      lobbyService,
		settlementService: settlementService,
		expenseService:    expenseService,
		userService:       userService,
		lastCheckDate:     time.Now(),
	}
}

// Start starts the scheduler in a goroutine
func (s *SettlementScheduler) Start() {
	go s.run()
}

// run runs the scheduler loop
func (s *SettlementScheduler) run() {
	// Check every hour
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	// Run initial check after a short delay
	time.Sleep(30 * time.Second)
	s.checkAndSendSettlements()

	for range ticker.C {
		s.checkAndSendSettlements()
	}
}

// checkAndSendSettlements checks if it's time to send monthly settlements
func (s *SettlementScheduler) checkAndSendSettlements() {
	now := time.Now()
	
	// Only check once per day, around midnight (between 00:00 and 01:00)
	if now.Hour() != 0 {
		return
	}

	// Check if we already processed today
	if now.Year() == s.lastCheckDate.Year() &&
		now.Month() == s.lastCheckDate.Month() &&
		now.Day() == s.lastCheckDate.Day() {
		return
	}

	// It's a new day, check if it's the 1st of the month (time to send last month's settlement)
	if now.Day() == 1 {
		s.sendMonthlySettlements(now)
		s.lastCheckDate = now
	}
}

// sendMonthlySettlements sends settlement reports for the previous month to all active lobbies
func (s *SettlementScheduler) sendMonthlySettlements(now time.Time) {
	// Calculate previous month
	prevMonth := now.AddDate(0, -1, 0)
	startDate, endDate := utils.GetMonthStartEnd(prevMonth.Year(), prevMonth.Month())

	log.Printf("Scheduler: Sending monthly settlements for %s", utils.FormatMonth(prevMonth))

	// Get all active lobbies (lobbies with both users)
	// We'll need to add a method to get all active lobbies
	// For now, we'll get lobbies by checking all users
	// This is a simplified approach - in production you might want a dedicated method

	// Get all lobbies (we'll need to add this method to lobbyService)
	lobbies, err := s.getAllActiveLobbies()
	if err != nil {
		log.Printf("Scheduler: Error getting lobbies: %v", err)
		return
	}

	for _, lobby := range lobbies {
		// Check if there are expenses for this month
		expenses, err := s.expenseService.GetExpensesByLobby(lobby.ID, &startDate, &endDate, nil)
		if err != nil {
			log.Printf("Scheduler: Error getting expenses for lobby %d: %v", lobby.ID, err)
			continue
		}

		if len(expenses) == 0 {
			// No expenses for this month, skip
			continue
		}

		// Calculate settlement
		result, err := s.settlementService.CalculateSettlement(lobby.ID, &startDate, &endDate)
		if err != nil {
			log.Printf("Scheduler: Error calculating settlement for lobby %d: %v", lobby.ID, err)
			continue
		}

		// Get translator (use first user's language preference)
		translator := s.getTranslatorForLobby(lobby)

		// Format settlement message
		// We need to format it - let's create a simple formatter
		msg := s.formatSettlementResultSimple(result, &startDate, &endDate, translator)
		
		// Add header indicating this is an automatic monthly settlement
		lang := translator.GetLanguage()
		var header string
		if lang == i18n.LanguageSpanishAR {
			header = "📅 *Resumen mensual de gastos*\n" + utils.FormatMonth(prevMonth) + "\n\n"
		} else {
			header = "📅 *Monthly Settlement Report*\n" + utils.FormatMonth(prevMonth) + "\n\n"
		}
		msg = header + msg

		// Send to both users
		// Try to get group chat ID if available
		if lobby.GroupChatID.Valid {
			// Send to group
			s.sendMessage(lobby.GroupChatID.Int64, msg)
		} else {
			// Send to both users in private
			s.sendMessage(lobby.User1TelegramID, msg)
			if lobby.User2TelegramID != 0 {
				s.sendMessage(lobby.User2TelegramID, msg)
			}
		}

		log.Printf("Scheduler: Sent monthly settlement to lobby %d", lobby.ID)
	}
}

// getAllActiveLobbies gets all active lobbies (with both users)
// This is a helper method - you might want to add this to lobbyService
func (s *SettlementScheduler) getAllActiveLobbies() ([]*database.Lobby, error) {
	// We'll need to add a method to lobbyService to get all active lobbies
	// For now, this is a placeholder that would need to be implemented
	// in the lobbyService
	return s.lobbyService.GetAllActiveLobbies()
}

// getTranslatorForLobby gets a translator for a lobby (uses first user's language)
func (s *SettlementScheduler) getTranslatorForLobby(lobby *database.Lobby) *i18n.Translator {
	// Try first user's language preference
	if s.userService != nil {
		user, err := s.userService.GetUserByTelegramID(lobby.User1TelegramID)
		if err == nil && user != nil && user.Language.Valid && user.Language.String != "" {
			return i18n.NewTranslator(i18n.Language(user.Language.String))
		}
	}
	return i18n.NewTranslator(i18n.LanguageEnglish)
}

// formatSettlementResultSimple formats a settlement result (with i18n support).
func (s *SettlementScheduler) formatSettlementResultSimple(result *service.SettlementResult, startDate, endDate *time.Time, translator *i18n.Translator) string {
	lang := translator.GetLanguage()
	isSpanish := lang == i18n.LanguageSpanishAR

	periodStr := translator.T("summary_period")
	if startDate != nil && endDate != nil {
		if isSpanish {
			periodStr = fmt.Sprintf("%s a %s", utils.FormatDate(*startDate), utils.FormatDate(*endDate))
		} else {
			periodStr = fmt.Sprintf("%s to %s", utils.FormatDate(*startDate), utils.FormatDate(*endDate))
		}
	}

	var msg string
	if isSpanish {
		msg = fmt.Sprintf("Período: %s\n", periodStr)
		msg += fmt.Sprintf("Tipo de cuenta: %s\n", result.AccountType)
		msg += fmt.Sprintf("Total de gastos: %s\n\n", utils.FormatCurrency(result.TotalExpenses))
	} else {
		msg = fmt.Sprintf("Period: %s\n", periodStr)
		msg += fmt.Sprintf("Account Type: %s\n", result.AccountType)
		msg += fmt.Sprintf("Total Expenses: %s\n\n", utils.FormatCurrency(result.TotalExpenses))
	}

	if isSpanish {
		msg += fmt.Sprintf("Usuario 1 gastó: %s\n", utils.FormatCurrency(result.User1TotalSpent))
		msg += fmt.Sprintf("Usuario 2 gastó: %s\n\n", utils.FormatCurrency(result.User2TotalSpent))
	} else {
		msg += fmt.Sprintf("User 1 Spent: %s\n", utils.FormatCurrency(result.User1TotalSpent))
		msg += fmt.Sprintf("User 2 Spent: %s\n\n", utils.FormatCurrency(result.User2TotalSpent))
	}

	if result.User1SalaryPercentage != 0.5 || result.User2SalaryPercentage != 0.5 || result.AccountType == "shared" {
		if isSpanish {
			msg += fmt.Sprintf("Esperado User 1 (%.1f%%): %s\n", result.User1SalaryPercentage*100, utils.FormatCurrency(result.User1Expected))
			msg += fmt.Sprintf("Esperado User 2 (%.1f%%): %s\n\n", result.User2SalaryPercentage*100, utils.FormatCurrency(result.User2Expected))
		} else {
			msg += fmt.Sprintf("User 1 Expected (%.1f%%): %s\n", result.User1SalaryPercentage*100, utils.FormatCurrency(result.User1Expected))
			msg += fmt.Sprintf("User 2 Expected (%.1f%%): %s\n\n", result.User2SalaryPercentage*100, utils.FormatCurrency(result.User2Expected))
		}
	} else {
		expectedPerPerson := result.TotalExpenses / 2.0
		if isSpanish {
			msg += fmt.Sprintf("Esperado por persona: %s\n\n", utils.FormatCurrency(expectedPerPerson))
		} else {
			msg += fmt.Sprintf("Expected per person: %s\n\n", utils.FormatCurrency(expectedPerPerson))
		}
	}

	if result.User1Debt > 0 {
		if isSpanish {
			msg += fmt.Sprintf("➡️ User 1 le debe a User 2: %s\n", utils.FormatCurrency(result.User1Debt))
		} else {
			msg += fmt.Sprintf("➡️ User 1 owes User 2: %s\n", utils.FormatCurrency(result.User1Debt))
		}
	} else if result.User1Debt < 0 {
		if isSpanish {
			msg += fmt.Sprintf("➡️ User 2 le debe a User 1: %s\n", utils.FormatCurrency(-result.User1Debt))
		} else {
			msg += fmt.Sprintf("➡️ User 2 owes User 1: %s\n", utils.FormatCurrency(-result.User1Debt))
		}
	} else {
		if isSpanish {
			msg += "✅ Todo saldado!\n"
		} else {
			msg += "✅ All settled!\n"
		}
	}

	return msg
}

// sendMessage sends a message (helper method)
func (s *SettlementScheduler) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeHTML
	if _, err := s.bot.Send(msg); err != nil {
		log.Printf("Scheduler: Error sending message to %d: %v", chatID, err)
	}
}

