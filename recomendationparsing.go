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

// ParseExpenseFromMessage parses an expense from a natural language message
func ParseExpenseFromMessage(ctx context.Context, provider LLMProvider, message string, context ExpenseContext) (*ParsedExpense, error) {
	return provider.ParseExpense(ctx, message, context)
}

// buildExpenseParsingPrompt builds a context-aware prompt for expense parsing
func buildExpenseParsingPrompt(message string, context ExpenseContext) string {
	var prompt strings.Builder

	// System role with clear instructions
	if context.UserLanguage == "es_AR" || strings.Contains(strings.ToLower(message), "gast") || strings.Contains(strings.ToLower(message), "compr") {
		prompt.WriteString(`Eres un asistente experto en interpretar gastos compartidos en español argentino. Tu trabajo es extraer información estructurada de mensajes informales sobre gastos.`)
	} else {
		prompt.WriteString(`You are an expert assistant for interpreting shared expenses. Your job is to extract structured information from informal messages about expenses.`)
	}

	prompt.WriteString("\n\n")

	// Context section with enhanced details
	prompt.WriteString("# CONTEXTO DEL USUARIO\n")
	prompt.WriteString(fmt.Sprintf("Usuario actual: %s (ID: %d)\n", context.UserDisplayName, context.UserID))
	
	if context.PartnerName != "" {
		prompt.WriteString(fmt.Sprintf("Pareja: %s\n", context.PartnerName))
	}
	
	if context.Lobby != nil {
		if context.Lobby.User1TelegramID == context.UserID {
			prompt.WriteString("Rol: Tú eres user1, tu pareja es user2\n")
		} else {
			prompt.WriteString("Rol: Tú eres user2, tu pareja es user1\n")
		}
		prompt.WriteString(fmt.Sprintf("Tipo de cuenta: %s\n", context.Lobby.AccountType))
	}

	// Payment methods with examples
	if len(context.PaymentMethods) > 0 {
		prompt.WriteString("\nMétodos de pago disponibles:\n")
		for _, pm := range context.PaymentMethods {
			if pm.IsActive {
				prompt.WriteString(fmt.Sprintf("- %s\n", pm.Name))
			}
		}
	}

	// Recent expenses for temporal context
	if len(context.RecentExpenses) > 0 {
		prompt.WriteString("\nGastos recientes (para referencia temporal y de patrones):\n")
		now := time.Now()
		for i, exp := range context.RecentExpenses {
			if i >= 5 {
				break
			}
			daysAgo := int(now.Sub(exp.ExpenseDate).Hours() / 24)
			desc := "sin descripción"
			if exp.Description.Valid {
				desc = exp.Description.String
			}
			cat := ""
			if exp.Category.Valid {
				cat = fmt.Sprintf(" [%s]", exp.Category.String)
			}
			pm := ""
			if exp.PaymentMethod.Valid {
				pm = fmt.Sprintf(" (%s)", exp.PaymentMethod.String)
			}
			
			var timeRef string
			switch daysAgo {
			case 0:
				timeRef = "hoy"
			case 1:
				timeRef = "ayer"
			default:
				timeRef = fmt.Sprintf("hace %d días", daysAgo)
			}
			
			prompt.WriteString(fmt.Sprintf("  %s: $%.2f - %s%s%s\n", timeRef, exp.Amount, desc, cat, pm))
		}
	}

	// Enhanced few-shot examples
	prompt.WriteString("\n# EJEMPLOS DE CONVERSIÓN\n")
	if context.UserLanguage == "es_AR" {
		prompt.WriteString(`
Entrada: "Gasté 20000 en helado con efectivo de Cami el 15 de diciembre"
Salida: {"amount": 20000, "description": "helado", "category": "comida", "payment_method": "efectivoCami", "spender": "me", "date": "2025-12-15", "confidence": 0.98, "needs_clarification": false, "clarification_question": ""}

Entrada: "compré $5000 de comida en el super"
Salida: {"amount": 5000, "description": "super", "category": "comida", "payment_method": "", "spender": "me", "date": "today", "confidence": 0.90, "needs_clarification": false, "clarification_question": ""}

Entrada: "mi pareja pagó 3500 de uber ayer con tarjeta"
Salida: {"amount": 3500, "description": "uber", "category": "transporte", "payment_method": "tarjeta", "spender": "partner", "date": "yesterday", "confidence": 0.92, "needs_clarification": false, "clarification_question": ""}

Entrada: "15000 luz"
Salida: {"amount": 15000, "description": "luz", "category": "servicios", "payment_method": "", "spender": "me", "date": "today", "confidence": 0.85, "needs_clarification": false, "clarification_question": ""}

Entrada: "Cami pagó el alquiler"
Salida: {"amount": 0, "description": "alquiler", "category": "vivienda", "payment_method": "", "spender": "partner", "date": "today", "confidence": 0.60, "needs_clarification": true, "clarification_question": "¿Cuánto pagó Cami por el alquiler?"}

Entrada: "compré comida"
Salida: {"amount": 0, "description": "comida", "category": "comida", "payment_method": "", "spender": "me", "date": "today", "confidence": 0.50, "needs_clarification": true, "clarification_question": "¿Cuánto gastaste en comida?"}

Entrada: "salimos a comer 8500 cada uno"
Salida: {"amount": 8500, "description": "salimos a comer", "category": "comida", "payment_method": "", "spender": "me", "date": "today", "confidence": 0.88, "needs_clarification": false, "clarification_question": "", "note": "Split expense, register for current user"}
`)
	} else {
		prompt.WriteString(`
Input: "I spent 20000 on ice cream with Cami's cash on December 15th"
Output: {"amount": 20000, "description": "ice cream", "category": "food", "payment_method": "cashCami", "spender": "me", "date": "2025-12-15", "confidence": 0.98, "needs_clarification": false, "clarification_question": ""}

Input: "bought $5000 of groceries at the supermarket"
Output: {"amount": 5000, "description": "supermarket", "category": "food", "payment_method": "", "spender": "me", "date": "today", "confidence": 0.90, "needs_clarification": false, "clarification_question": ""}

Input: "my partner paid 3500 for uber yesterday with card"
Output: {"amount": 3500, "description": "uber", "category": "transport", "payment_method": "card", "spender": "partner", "date": "yesterday", "confidence": 0.92, "needs_clarification": false, "clarification_question": ""}

Input: "15000 electricity"
Output: {"amount": 15000, "description": "electricity", "category": "utilities", "payment_method": "", "spender": "me", "date": "today", "confidence": 0.85, "needs_clarification": false, "clarification_question": ""}

Input: "Cami paid rent"
Output: {"amount": 0, "description": "rent", "category": "housing", "payment_method": "", "spender": "partner", "date": "today", "confidence": 0.60, "needs_clarification": true, "clarification_question": "How much did Cami pay for rent?"}

Input: "bought food"
Output: {"amount": 0, "description": "food", "category": "food", "payment_method": "", "spender": "me", "date": "today", "confidence": 0.50, "needs_clarification": true, "clarification_question": "How much did you spend on food?"}
`)
	}

	// Rules and guidelines
	prompt.WriteString("\n# REGLAS DE INTERPRETACIÓN\n")
	if context.UserLanguage == "es_AR" {
		prompt.WriteString(`
1. CANTIDAD: Extraer números que representen dinero. Si no hay cantidad, amount = 0 y needs_clarification = true
2. DESCRIPCIÓN: Breve descripción del gasto (ej: "super", "uber", "helado")
3. CATEGORÍA: Inferir categoría inteligente:
   - comida: supermercado, restaurant, delivery, helado, panadería
   - transporte: uber, taxi, nafta, peaje, colectivo
   - servicios: luz, gas, agua, internet, teléfono
   - vivienda: alquiler, expensas
   - salud: farmacia, médico, prepaga
   - entretenimiento: cine, streaming, salidas
   - otros: todo lo demás
4. MÉTODO DE PAGO: Buscar coincidencias con los métodos disponibles o inferir (efectivo, tarjeta, transferencia)
5. PAGADOR (spender):
   - "yo", "gasté", "pagué", "compré" → "me"
   - "mi pareja", "Cami", "partner" → "partner"
   - Si no está claro → "me" (default)
6. FECHA: 
   - "hoy", "today" → "today"
   - "ayer", "yesterday" → "yesterday"
   - Fecha específica → formato YYYY-MM-DD
   - "anteayer" → calcular fecha hace 2 días
   - "hace X días" → calcular fecha relativa
   - Sin mención → "today"
7. CONFIANZA (confidence): 
   - 0.9-1.0: Toda la información está clara
   - 0.7-0.89: Información mayormente clara, algunas inferencias
   - 0.5-0.69: Información parcial, necesita confirmación
   - 0.0-0.49: Información muy ambigua
8. CLARIFICACIÓN: Si falta información crítica (especialmente monto), needs_clarification = true y generar pregunta específica
`)
	} else {
		prompt.WriteString(`
1. AMOUNT: Extract numbers representing money. If no amount, amount = 0 and needs_clarification = true
2. DESCRIPTION: Brief description of expense (e.g., "groceries", "uber", "ice cream")
3. CATEGORY: Infer smart category:
   - food: supermarket, restaurant, delivery, groceries
   - transport: uber, taxi, gas, tolls, bus
   - utilities: electricity, gas, water, internet, phone
   - housing: rent, HOA fees
   - health: pharmacy, doctor, insurance
   - entertainment: movies, streaming, outings
   - other: everything else
4. PAYMENT METHOD: Match with available methods or infer (cash, card, transfer)
5. SPENDER:
   - "I", "spent", "paid", "bought" → "me"
   - "my partner", "partner name" → "partner"
   - If unclear → "me" (default)
6. DATE:
   - "today" → "today"
   - "yesterday" → "yesterday"
   - Specific date → YYYY-MM-DD format
   - "X days ago" → calculate relative date
   - No mention → "today"
7. CONFIDENCE:
   - 0.9-1.0: All information is clear
   - 0.7-0.89: Mostly clear, some inference
   - 0.5-0.69: Partial info, needs confirmation
   - 0.0-0.49: Very ambiguous
8. CLARIFICATION: If critical info missing (especially amount), needs_clarification = true and generate specific question
`)
	}

	// Current message to parse
	prompt.WriteString("\n# MENSAJE A PARSEAR\n")
	prompt.WriteString(fmt.Sprintf("Mensaje del usuario: \"%s\"\n\n", message))

	// Output format specification
	prompt.WriteString("# FORMATO DE SALIDA REQUERIDO\n")
	prompt.WriteString("Devuelve ÚNICAMENTE un objeto JSON válido (sin markdown, sin explicaciones, sin bloques de código):\n\n")
	prompt.WriteString(`{"amount": <number>, "description": "<string>", "category": "<string>", "payment_method": "<string>", "spender": "me|partner", "date": "YYYY-MM-DD|today|yesterday", "confidence": <0.0-1.0>, "needs_clarification": <boolean>, "clarification_question": "<string>"}`)

	return prompt.String()
}