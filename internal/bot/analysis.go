package bot

import (
	"botGastosPareja/internal/service"
	"botGastosPareja/pkg/utils"
	"fmt"
	"sort"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// registerAnalysisCommands registers analysis-related commands
func (h *Handler) registerAnalysisCommands() {
	h.router.RegisterCommand("analyze", h.handleAnalyze)
}

// handleAnalyze handles the /analyze command
func (h *Handler) handleAnalyze(handler *Handler, message *tgbotapi.Message, args string) {
	// Get user's lobby for this specific chat (group/private)
	lobby, err := handler.getLobbyForMessage(message)
	if err != nil || lobby == nil {
		handler.sendMessage(message.Chat.ID,
			"❌ You're not in a lobby yet. Use /start to create or join one.")
		return
	}

	result, err := handler.analysisService.AnalyzeMonthly(lobby.ID)
	if err != nil {
		handler.sendMessage(message.Chat.ID,
			fmt.Sprintf("❌ Error analyzing spending: %v", err))
		return
	}

	msg := h.formatAnalysisResult(result)
	handler.sendMessage(message.Chat.ID, msg)
}

// formatAnalysisResult formats analysis results for display
func (h *Handler) formatAnalysisResult(result *service.AnalysisResult) string {
	msg := fmt.Sprintf("📈 *Monthly Spending Analysis*\n\n"+
		"Current Period: %s\n"+
		"Previous Period: %s\n\n",
		utils.FormatMonth(result.CurrentPeriod),
		utils.FormatMonth(result.PreviousPeriod))

	// Overall comparison
	msg += fmt.Sprintf("*Overall Spending:*\n"+
		"Current: %s\n"+
		"Previous: %s\n",
		utils.FormatCurrency(result.CurrentTotal),
		utils.FormatCurrency(result.PreviousTotal))

	if result.ChangePercent > 0 {
		msg += fmt.Sprintf("📈 Increase: %.1f%%\n\n", result.ChangePercent)
	} else if result.ChangePercent < 0 {
		msg += fmt.Sprintf("📉 Decrease: %.1f%%\n\n", -result.ChangePercent)
	} else {
		msg += "➡️ No change\n\n"
	}

	// Spending spikes
	if len(result.SpendingSpikes) > 0 {
		msg += "*⚠️ Spending Spikes (>20%% increase):*\n"
		for _, spike := range result.SpendingSpikes {
			msg += fmt.Sprintf("• %s: %s (+%.1f%%)\n",
				spike.Category,
				utils.FormatCurrency(spike.Amount),
				spike.ChangePercent)
		}
		msg += "\n"
	}

	// New categories
	if len(result.NewCategories) > 0 {
		msg += "*🆕 New Categories:*\n"
		for _, cat := range result.NewCategories {
			msg += fmt.Sprintf("• %s\n", cat)
		}
		msg += "\n"
	}

	// Discontinued categories
	if len(result.DiscontinuedCategories) > 0 {
		msg += "*❌ Discontinued Categories:*\n"
		for _, cat := range result.DiscontinuedCategories {
			msg += fmt.Sprintf("• %s\n", cat)
		}
		msg += "\n"
	}

	// Top category changes
	if len(result.CategoryChanges) > 0 {
		type catChange struct {
			name   string
			change service.CategoryChange
		}
		changes := make([]catChange, 0, len(result.CategoryChanges))
		for name, change := range result.CategoryChanges {
			changes = append(changes, catChange{name, change})
		}
		sort.Slice(changes, func(i, j int) bool {
			return changes[i].change.ChangePercent > changes[j].change.ChangePercent
		})

		msg += "*Top Category Changes:*\n"
		count := 0
		for _, cc := range changes {
			if count >= 5 {
				break
			}
			if cc.change.PreviousTotal > 0 {
				msg += fmt.Sprintf("• %s: %s → %s (%.1f%%)\n",
					cc.name,
					utils.FormatCurrency(cc.change.PreviousTotal),
					utils.FormatCurrency(cc.change.CurrentTotal),
					cc.change.ChangePercent)
				count++
			}
		}
	}

	return msg
}
