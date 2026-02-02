//go:build ignore
// +build ignore

package llm

import (
	"botGastosPareja/internal/database"
	"context"
	"fmt"
	"strings"
	"time"
)

// GenerateAnalysisInsights generates AI-powered insights about expenses
func GenerateAnalysisInsights(ctx context.Context, provider LLMProvider, expenses []*database.Expense, lobby *database.Lobby, analysisType string) (*AnalysisInsights, error) {
	return provider.AnalyzeExpenses(ctx, expenses, lobby, analysisType)
}

// buildAnalysisPrompt builds a prompt for expense analysis
func buildAnalysisPrompt(expenses []*database.Expense, lobby *database.Lobby, analysisType string) string {
	var prompt strings.Builder

	// System role
	prompt.WriteString("Eres un asistente financiero experto que analiza gastos compartidos de parejas. ")
	prompt.WriteString("Tu objetivo es proporcionar insights accionables, identificar patrones, y ayudar a mejorar la gestión financiera.\n\n")

	// Lobby context
	prompt.WriteString("# CONTEXTO DE LA CUENTA\n")
	prompt.WriteString(fmt.Sprintf("Tipo de cuenta: %s\n", lobby.AccountType))

	if lobby.AccountType == "shared" {
		prompt.WriteString(fmt.Sprintf("Distribución por sueldo:\n"))
		prompt.WriteString(fmt.Sprintf("- User1: %.1f%% del total\n", lobby.User1SalaryPercentage*100))
		prompt.WriteString(fmt.Sprintf("- User2: %.1f%% del total\n", lobby.User2SalaryPercentage*100))
		prompt.WriteString("(Cada persona debería pagar según su porcentaje de ingreso)\n")
	} else {
		prompt.WriteString("División: 50/50 (Cada persona debería pagar la mitad)\n")
	}

	// Time context
	now := time.Now()
	currentMonthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	previousMonthStart := currentMonthStart.AddDate(0, -1, 0)
	lastMonthStart := previousMonthStart.AddDate(0, -1, 0)

	prompt.WriteString(fmt.Sprintf("\nFecha actual: %s\n", now.Format("2006-01-02")))
	prompt.WriteString(fmt.Sprintf("Mes actual: %s\n", currentMonthStart.Format("January 2006")))

	// Aggregate data
	var total float64
	categoryTotals := make(map[string]float64)
	paymentMethodTotals := make(map[string]float64)
	user1Total := 0.0
	user2Total := 0.0

	currentMonthExpenses := make([]*database.Expense, 0)
	previousMonthExpenses := make([]*database.Expense, 0)
	lastMonthExpenses := make([]*database.Expense, 0)

	categoryCount := make(map[string]int)
	dailyTotals := make(map[string]float64)
	weekdayTotals := make(map[time.Weekday]float64)

	for _, exp := range expenses {
		total += exp.Amount

		// User totals
		if exp.SpenderTelegramID == lobby.User1TelegramID {
			user1Total += exp.Amount
		} else if exp.SpenderTelegramID == lobby.User2TelegramID {
			user2Total += exp.Amount
		}

		// Category aggregation
		if exp.Category.Valid {
			categoryTotals[exp.Category.String] += exp.Amount
			categoryCount[exp.Category.String]++
		}

		// Payment method aggregation
		if exp.PaymentMethod.Valid {
			paymentMethodTotals[exp.PaymentMethod.String] += exp.Amount
		}

		// Monthly grouping
		if exp.ExpenseDate.After(currentMonthStart) || exp.ExpenseDate.Equal(currentMonthStart) {
			currentMonthExpenses = append(currentMonthExpenses, exp)
		} else if exp.ExpenseDate.After(previousMonthStart) && exp.ExpenseDate.Before(currentMonthStart) {
			previousMonthExpenses = append(previousMonthExpenses, exp)
		} else if exp.ExpenseDate.After(lastMonthStart) && exp.ExpenseDate.Before(previousMonthStart) {
			lastMonthExpenses = append(lastMonthExpenses, exp)
		}

		// Daily and weekday patterns
		dayKey := exp.ExpenseDate.Format("2006-01-02")
		dailyTotals[dayKey] += exp.Amount
		weekdayTotals[exp.ExpenseDate.Weekday()] += exp.Amount
	}

	// Summary statistics
	prompt.WriteString("\n# RESUMEN ESTADÍSTICO\n")
	prompt.WriteString(fmt.Sprintf("Total de gastos: $%.2f\n", total))
	prompt.WriteString(fmt.Sprintf("Cantidad de gastos: %d\n", len(expenses)))
	if len(expenses) > 0 {
		prompt.WriteString(fmt.Sprintf("Gasto promedio: $%.2f\n", total/float64(len(expenses))))
	}

	prompt.WriteString("\nPor persona:\n")
	prompt.WriteString(fmt.Sprintf("- User1: $%.2f (%.1f%%)\n", user1Total, (user1Total/total)*100))
	prompt.WriteString(fmt.Sprintf("- User2: $%.2f (%.1f%%)\n", user2Total, (user2Total/total)*100))

	// Expected vs actual
	if lobby.AccountType == "shared" {
		expectedUser1 := total * lobby.User1SalaryPercentage
		expectedUser2 := total * lobby.User2SalaryPercentage
		diff1 := user1Total - expectedUser1
		diff2 := user2Total - expectedUser2

		prompt.WriteString("\nEsperado vs Real:\n")
		prompt.WriteString(fmt.Sprintf("- User1: Esperado $%.2f, Gastó $%.2f (diferencia: $%.2f)\n", expectedUser1, user1Total, diff1))
		prompt.WriteString(fmt.Sprintf("- User2: Esperado $%.2f, Gastó $%.2f (diferencia: $%.2f)\n", expectedUser2, user2Total, diff2))
	}

	// Monthly comparison
	currentMonthTotal := 0.0
	for _, e := range currentMonthExpenses {
		currentMonthTotal += e.Amount
	}
	previousMonthTotal := 0.0
	for _, e := range previousMonthExpenses {
		previousMonthTotal += e.Amount
	}
	lastMonthTotal := 0.0
	for _, e := range lastMonthExpenses {
		lastMonthTotal += e.Amount
	}

	prompt.WriteString("\nComparación mensual:\n")
	prompt.WriteString(fmt.Sprintf("- Mes actual: $%.2f (%d gastos)\n", currentMonthTotal, len(currentMonthExpenses)))
	prompt.WriteString(fmt.Sprintf("- Mes anterior: $%.2f (%d gastos)\n", previousMonthTotal, len(previousMonthExpenses)))
	if previousMonthTotal > 0 {
		change := ((currentMonthTotal - previousMonthTotal) / previousMonthTotal) * 100
		prompt.WriteString(fmt.Sprintf("- Cambio: %.1f%%\n", change))
	}
	if len(lastMonthExpenses) > 0 {
		prompt.WriteString(fmt.Sprintf("- Hace 2 meses: $%.2f (%d gastos)\n", lastMonthTotal, len(lastMonthExpenses)))
	}

	// Category breakdown
	if len(categoryTotals) > 0 {
		prompt.WriteString("\nDistribución por categoría:\n")
		for cat, amt := range categoryTotals {
			percentage := (amt / total) * 100
			avgPerExpense := amt / float64(categoryCount[cat])
			prompt.WriteString(fmt.Sprintf("- %s: $%.2f (%.1f%%, %d gastos, promedio $%.2f)\n",
				cat, amt, percentage, categoryCount[cat], avgPerExpense))
		}
	}

	// Payment method breakdown
	if len(paymentMethodTotals) > 0 {
		prompt.WriteString("\nPor método de pago:\n")
		for pm, amt := range paymentMethodTotals {
			percentage := (amt / total) * 100
			prompt.WriteString(fmt.Sprintf("- %s: $%.2f (%.1f%%)\n", pm, amt, percentage))
		}
	}

	// Temporal patterns
	prompt.WriteString("\nPatrones temporales:\n")
	maxDay := ""
	maxAmount := 0.0
	for day, amt := range dailyTotals {
		if amt > maxAmount {
			maxAmount = amt
			maxDay = day
		}
	}
	if maxDay != "" {
		prompt.WriteString(fmt.Sprintf("- Día con más gastos: %s ($%.2f)\n", maxDay, maxAmount))
	}

	// Weekday analysis
	for wd := time.Sunday; wd <= time.Saturday; wd++ {
		if amt, ok := weekdayTotals[wd]; ok && amt > 0 {
			prompt.WriteString(fmt.Sprintf("- %s: $%.2f\n", wd.String(), amt))
		}
	}

	// Recent expenses detail
	prompt.WriteString("\n# GASTOS RECIENTES (últimos 15)\n")
	count := 0
	for _, exp := range expenses {
		if count >= 15 {
			break
		}
		desc := "sin descripción"
		if exp.Description.Valid {
			desc = exp.Description.String
		}
		cat := "sin categoría"
		if exp.Category.Valid {
			cat = exp.Category.String
		}
		pm := "sin método"
		if exp.PaymentMethod.Valid {
			pm = exp.PaymentMethod.String
		}

		spender := "User1"
		if exp.SpenderTelegramID == lobby.User2TelegramID {
			spender = "User2"
		}

		daysAgo := int(now.Sub(exp.ExpenseDate).Hours() / 24)
		timeRef := fmt.Sprintf("hace %d días", daysAgo)
		if daysAgo == 0 {
			timeRef = "hoy"
		} else if daysAgo == 1 {
			timeRef = "ayer"
		}

		prompt.WriteString(fmt.Sprintf("%s (%s): $%.2f - %s [%s] (%s) - %s\n",
			exp.ExpenseDate.Format("2006-01-02"), timeRef, exp.Amount, desc, cat, pm, spender))
		count++
	}

	// Analysis type
	prompt.WriteString(fmt.Sprintf("\n# TIPO DE ANÁLISIS: %s\n", analysisType))

	// Analysis instructions
	prompt.WriteString("\n# INSTRUCCIONES DE ANÁLISIS\n")
	prompt.WriteString(`
Analiza los datos y genera insights en las siguientes áreas:

1. NARRATIVA (narrative): 
   - Resumen general en lenguaje natural (2-3 párrafos)
   - Destacar tendencias principales
   - Mencionar equilibrio entre usuarios
   - Comentar sobre la salud financiera general

2. PATRONES (patterns):
   - Tendencias de gasto (aumentos/disminuciones)
   - Cambios en categorías específicas
   - Velocidad de gasto (aceleración/desaceleración)
   - Patrones de comportamiento recurrentes

3. ANOMALÍAS (anomalies):
   - Gastos inusuales por monto
   - Posibles duplicados
   - Gastos fuera de patrón temporal
   - Gastos sin categoría o método

4. RECOMENDACIONES (recommendations):
   - Sugerencias de presupuesto
   - Optimización de categorías
   - Mejoras en timing de gastos
   - Balanceo entre usuarios

5. INSIGHTS CONDUCTUALES (behavioral_insights):
   - Análisis de velocidad de gasto
   - Comparación entre usuarios (sin juicio)
   - Mix de categorías
   - Patrones de días/horarios
`)

	// Output format
	prompt.WriteString("\n# FORMATO DE SALIDA\n")
	prompt.WriteString("Devuelve ÚNICAMENTE un objeto JSON válido (sin markdown, sin explicaciones):\n\n")
	prompt.WriteString(`{
  "narrative": "<resumen natural de 2-3 párrafos sobre patrones de gasto, equilibrio entre usuarios, y salud financiera>",
  "patterns": [
    {
      "type": "trend|category_change|velocity|seasonal",
      "description": "<descripción clara del patrón>",
      "category": "<categoría si aplica>",
      "amount": <monto relevante>,
      "change": <porcentaje de cambio>,
      "time_period": "<período del patrón>"
    }
  ],
  "anomalies": [
    {
      "type": "unusual_amount|duplicate|temporal|missing_data",
      "description": "<descripción de la anomalía>",
      "expense_id": <id del gasto o 0>,
      "severity": "low|medium|high",
      "suggestion": "<sugerencia para resolver>"
    }
  ],
  "recommendations": [
    {
      "type": "budget|category|timing|balance",
      "description": "<descripción de la recomendación>",
      "action": "<acción específica y accionable>",
      "priority": "low|medium|high",
      "expected_impact": "<impacto esperado>"
    }
  ],
  "behavioral_insights": [
    {
      "type": "velocity|partner_comparison|category_mix|temporal_pattern",
      "description": "<insight conductual>",
      "data": {
        "metric": "<métrica relevante>",
        "value": <valor numérico>
      }
    }
  ],
  "summary_stats": {
    "total_expenses": <total>,
    "expense_count": <cantidad>,
    "average_expense": <promedio>,
    "user1_percentage": <porcentaje>,
    "user2_percentage": <porcentaje>,
    "top_category": "<categoría principal>",
    "monthly_trend": "increasing|decreasing|stable"
  }
}`)

	prompt.WriteString("\n\nIMPORTANTE:\n")
	prompt.WriteString("- Sé específico con números y porcentajes\n")
	prompt.WriteString("- Mantén un tono constructivo y no crítico\n")
	prompt.WriteString("- Enfócate en insights accionables\n")
	prompt.WriteString("- Usa lenguaje claro y amigable\n")
	prompt.WriteString("- Considera el contexto de pareja (shared account)\n")
	prompt.WriteString("- Identifica oportunidades de ahorro sin ser negativo\n")

	return prompt.String()
}
