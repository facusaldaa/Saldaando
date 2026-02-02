package llm

import (
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
	if context.UserLanguage == "es_AR" {
		prompt.WriteString("# CONTEXTO DEL USUARIO\n")
		prompt.WriteString(fmt.Sprintf("Usuario actual: %s (ID: %d)\n", context.UserDisplayName, context.UserID))
	} else {
		prompt.WriteString("# USER CONTEXT\n")
		prompt.WriteString(fmt.Sprintf("Current user: %s (ID: %d)\n", context.UserDisplayName, context.UserID))
	}
	
	if context.PartnerName != "" {
		if context.UserLanguage == "es_AR" {
			prompt.WriteString(fmt.Sprintf("Pareja: %s\n", context.PartnerName))
		} else {
			prompt.WriteString(fmt.Sprintf("Partner: %s\n", context.PartnerName))
		}
	}
	
	if context.Lobby != nil {
		if context.UserLanguage == "es_AR" {
			if context.Lobby.User1TelegramID == context.UserID {
				prompt.WriteString("Rol: Tú eres user1, tu pareja es user2\n")
			} else {
				prompt.WriteString("Rol: Tú eres user2, tu pareja es user1\n")
			}
			prompt.WriteString(fmt.Sprintf("Tipo de cuenta: %s\n", context.Lobby.AccountType))
		} else {
			if context.Lobby.User1TelegramID == context.UserID {
				prompt.WriteString("Role: You are user1, your partner is user2\n")
			} else {
				prompt.WriteString("Role: You are user2, your partner is user1\n")
			}
			prompt.WriteString(fmt.Sprintf("Account type: %s\n", context.Lobby.AccountType))
		}
	}

	// Payment methods with examples
	if len(context.PaymentMethods) > 0 {
		if context.UserLanguage == "es_AR" {
			prompt.WriteString("\nMétodos de pago disponibles:\n")
		} else {
			prompt.WriteString("\nAvailable payment methods:\n")
		}
		for _, pm := range context.PaymentMethods {
			if pm.IsActive {
				prompt.WriteString(fmt.Sprintf("- %s\n", pm.Name))
			}
		}
	}

	// Recent expenses for temporal context (enhanced to 5)
	if len(context.RecentExpenses) > 0 {
		if context.UserLanguage == "es_AR" {
			prompt.WriteString("\nGastos recientes (para referencia temporal y de patrones):\n")
		} else {
			prompt.WriteString("\nRecent expenses (for temporal and pattern reference):\n")
		}
		now := time.Now()
		for i, exp := range context.RecentExpenses {
			if i >= 5 {
				break
			}
			daysAgo := int(now.Sub(exp.ExpenseDate).Hours() / 24)
			desc := "sin descripción"
			if context.UserLanguage != "es_AR" {
				desc = "no description"
			}
			if exp.Description.Valid {
				desc = exp.Description.String
			}
			cat := ""
			if exp.Category.Valid {
				cat = fmt.Sprintf(" [%s]", exp.Category.String)
			}
			
			var timeRef string
			if context.UserLanguage == "es_AR" {
				switch daysAgo {
				case 0:
					timeRef = "hoy"
				case 1:
					timeRef = "ayer"
				default:
					timeRef = fmt.Sprintf("hace %d días", daysAgo)
				}
			} else {
				switch daysAgo {
				case 0:
					timeRef = "today"
				case 1:
					timeRef = "yesterday"
				default:
					timeRef = fmt.Sprintf("%d days ago", daysAgo)
				}
			}
			
			prompt.WriteString(fmt.Sprintf("  %s: $%.2f - %s%s\n", timeRef, exp.Amount, desc, cat))
		}
	}

	// Enhanced few-shot examples
	if context.UserLanguage == "es_AR" {
		prompt.WriteString("\n# EJEMPLOS DE CONVERSIÓN\n")
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
Salida: {"amount": 8500, "description": "salimos a comer", "category": "comida", "payment_method": "", "spender": "me", "date": "today", "confidence": 0.88, "needs_clarification": false, "clarification_question": "", "is_installment": false, "installments": 0}

Entrada: "compré un celular de 120000 en 12 cuotas con visa"
Salida: {"amount": 120000, "description": "celular", "category": "otros", "payment_method": "visa", "spender": "me", "date": "today", "confidence": 0.95, "needs_clarification": false, "clarification_question": "", "is_installment": true, "installments": 12}

Entrada: "zapatillas 30000 en 3 cuotas con mastercard"
Salida: {"amount": 30000, "description": "zapatillas", "category": "otros", "payment_method": "mastercard", "spender": "me", "date": "today", "confidence": 0.93, "needs_clarification": false, "clarification_question": "", "is_installment": true, "installments": 3}
`)
	} else {
		prompt.WriteString("\n# CONVERSION EXAMPLES\n")
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
Output: {"amount": 0, "description": "food", "category": "food", "payment_method": "", "spender": "me", "date": "today", "confidence": 0.50, "needs_clarification": true, "clarification_question": "How much did you spend on food?", "is_installment": false, "installments": 0}

Input: "bought a phone for 120000 in 12 installments with visa"
Output: {"amount": 120000, "description": "phone", "category": "other", "payment_method": "visa", "spender": "me", "date": "today", "confidence": 0.95, "needs_clarification": false, "clarification_question": "", "is_installment": true, "installments": 12}

Input: "shoes 30000 in 3 installments with mastercard"
Output: {"amount": 30000, "description": "shoes", "category": "other", "payment_method": "mastercard", "spender": "me", "date": "today", "confidence": 0.93, "needs_clarification": false, "clarification_question": "", "is_installment": true, "installments": 3}
`)
	}

	// Rules and guidelines
	if context.UserLanguage == "es_AR" {
		prompt.WriteString("\n# REGLAS DE INTERPRETACIÓN\n")
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
8. CUOTAS/INSTALLMENTS: Detectar si el gasto es en cuotas:
   - Palabras clave: "cuotas", "cuota", "en X cuotas", "X cuotas", "installments", "in X installments"
   - Si se menciona cuotas, is_installment = true y installments = número de cuotas mencionado
   - Si no se menciona, is_installment = false y installments = 0
   - Ejemplos: "12 cuotas" → installments = 12, "en 3 cuotas" → installments = 3
9. CLARIFICACIÓN: Si falta información crítica (especialmente monto), needs_clarification = true y generar pregunta específica
`)
	} else {
		prompt.WriteString("\n# INTERPRETATION RULES\n")
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
8. INSTALLMENTS: Detect if expense is in installments:
   - Keywords: "cuotas", "cuota", "en X cuotas", "X cuotas", "installments", "in X installments"
   - If installments mentioned, is_installment = true and installments = number mentioned
   - If not mentioned, is_installment = false and installments = 0
   - Examples: "12 cuotas" → installments = 12, "in 3 installments" → installments = 3
9. CLARIFICATION: If critical info missing (especially amount), needs_clarification = true and generate specific question
`)
	}

	// Current message to parse
	if context.UserLanguage == "es_AR" {
		prompt.WriteString("\n# MENSAJE A PARSEAR\n")
		prompt.WriteString(fmt.Sprintf("Mensaje del usuario: \"%s\"\n\n", message))
	} else {
		prompt.WriteString("\n# MESSAGE TO PARSE\n")
		prompt.WriteString(fmt.Sprintf("User message: \"%s\"\n\n", message))
	}

	// Output format specification
	if context.UserLanguage == "es_AR" {
		prompt.WriteString("# FORMATO DE SALIDA REQUERIDO\n")
		prompt.WriteString("Devuelve ÚNICAMENTE un objeto JSON válido (sin markdown, sin explicaciones, sin bloques de código):\n\n")
	} else {
		prompt.WriteString("# REQUIRED OUTPUT FORMAT\n")
		prompt.WriteString("Return ONLY a valid JSON object (no markdown, no explanations, no code blocks):\n\n")
	}
	prompt.WriteString(`{"amount": <number>, "description": "<string>", "category": "<string>", "payment_method": "<string>", "spender": "me|partner", "date": "YYYY-MM-DD|today|yesterday", "confidence": <0.0-1.0>, "needs_clarification": <boolean>, "clarification_question": "<string>", "is_installment": <boolean>, "installments": <number>}`)

	return prompt.String()
}
