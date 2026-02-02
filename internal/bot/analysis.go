package bot

import (
	"botGastosPareja/internal/service"
	"botGastosPareja/pkg/i18n"
	"botGastosPareja/pkg/utils"
	"fmt"
	"sort"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// registerAnalysisCommands registers analysis-related commands
func (h *Handler) registerAnalysisCommands() {
	h.router.RegisterCommand("analyze", h.handleAnalyze)
	// /review is AI-powered; only register if LLM features are enabled
	if h.llmEnabled {
		h.router.RegisterCommand("review", h.handleReview)
	}
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

// handleReview handles the /review command for AI-powered expense reviews
func (h *Handler) handleReview(handler *Handler, message *tgbotapi.Message, args string) {
	userID := message.From.ID
	translator := handler.getTranslator(userID)

	// Get user's lobby for this specific chat (group/private)
	lobby, err := handler.getLobbyForMessage(message)
	if err != nil || lobby == nil {
		handler.sendTranslatedMessage(userID, message.Chat.ID, "error_lobby_not_found")
		return
	}

	// Check if LLM service is available
	if handler.analysisService == nil {
		handler.sendMessage(message.Chat.ID, "❌ AI analysis is not available. Please configure LLM service.")
		return
	}

	// Get AI analysis
	result, err := handler.analysisService.AnalyzeWithAI(lobby.ID)
	if err != nil {
		handler.sendMessage(message.Chat.ID,
			fmt.Sprintf("❌ Error analyzing expenses: %v", err))
		return
	}

	msg := handler.formatAIAnalysisResult(result, translator)
	handler.sendMessage(message.Chat.ID, msg)
}

// formatAIAnalysisResult formats AI analysis results for display
func (h *Handler) formatAIAnalysisResult(result *service.AIAnalysisResult, translator *i18n.Translator) string {
	msg := fmt.Sprintf("🤖 *AI-Powered Expense Review*\n\n")

	// Basic analysis summary
	msg += fmt.Sprintf("📊 *Monthly Summary*\n")
	msg += fmt.Sprintf("Current Period: %s\n", utils.FormatMonth(result.CurrentPeriod))
	msg += fmt.Sprintf("Previous Period: %s\n\n", utils.FormatMonth(result.PreviousPeriod))

	msg += fmt.Sprintf("*Overall Spending:*\n")
	msg += fmt.Sprintf("Current: %s\n", utils.FormatCurrency(result.CurrentTotal))
	msg += fmt.Sprintf("Previous: %s\n", utils.FormatCurrency(result.PreviousTotal))

	if result.ChangePercent > 0 {
		msg += fmt.Sprintf("📈 Increase: %.1f%%\n\n", result.ChangePercent)
	} else if result.ChangePercent < 0 {
		msg += fmt.Sprintf("📉 Decrease: %.1f%%\n\n", -result.ChangePercent)
	} else {
		msg += "➡️ No change\n\n"
	}

	// AI Insights
	if result.AIInsights != nil {
		if result.AIInsights.Narrative != "" {
			msg += fmt.Sprintf("*💡 AI Insights:*\n%s\n\n", result.AIInsights.Narrative)
		}

		// Patterns
		if len(result.AIInsights.Patterns) > 0 {
			msg += "*📈 Detected Patterns:*\n"
			for _, pattern := range result.AIInsights.Patterns {
				msg += fmt.Sprintf("• %s", pattern.Description)
				if pattern.Category != "" {
					msg += fmt.Sprintf(" (%s)", pattern.Category)
				}
				if pattern.Change != 0 {
					msg += fmt.Sprintf(" - %.1f%% change", pattern.Change)
				}
				msg += "\n"
			}
			msg += "\n"
		}

		// Anomalies
		if len(result.AIInsights.Anomalies) > 0 {
			msg += "*⚠️ Anomalies Detected:*\n"
			for _, anomaly := range result.AIInsights.Anomalies {
				severityIcon := "🔵"
				if anomaly.Severity == "high" {
					severityIcon = "🔴"
				} else if anomaly.Severity == "medium" {
					severityIcon = "🟡"
				}
				msg += fmt.Sprintf("%s %s\n", severityIcon, anomaly.Description)
			}
			msg += "\n"
		}

		// Recommendations
		if len(result.AIInsights.Recommendations) > 0 {
			msg += "*💼 Recommendations:*\n"
			for _, rec := range result.AIInsights.Recommendations {
				priorityIcon := "💡"
				if rec.Priority == "high" {
					priorityIcon = "🔥"
				} else if rec.Priority == "medium" {
					priorityIcon = "⭐"
				}
				msg += fmt.Sprintf("%s %s\n", priorityIcon, rec.Description)
				if rec.Action != "" {
					msg += fmt.Sprintf("   → %s\n", rec.Action)
				}
			}
			msg += "\n"
		}

		// Behavioral insights
		if len(result.AIInsights.BehavioralInsights) > 0 {
			msg += "*🧠 Behavioral Insights:*\n"
			for _, insight := range result.AIInsights.BehavioralInsights {
				msg += fmt.Sprintf("• %s\n", insight.Description)
			}
			msg += "\n"
		}
	} else {
		msg += "\n*Note:* AI insights are not available. Basic analysis only.\n"
	}

	return msg
}
