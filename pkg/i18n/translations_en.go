package i18n

var englishTranslations = map[string]string{
	// Welcome messages
	"welcome":             "👋 Welcome to the Couple Expense Tracker Bot!\n\nHi %s! I'll help you and your partner track shared expenses.\n\nTo get started:\n1. Create a lobby with /start\n2. Share the lobby code with your partner\n3. Start adding expenses with /add\n\nUse /help to see all available commands.",
	"welcome_back":        "👋 Welcome back, %s!\n\n*Your Lobby:*\nLobby ID: `%d`\nAccount Type: %s\n%s\n\nUse /help to see all available commands.",
	"lobby_created":       "👋 Welcome to the Couple Expense Tracker Bot!\n\nHi %s! I've created a new lobby for you.\n\n*Your Lobby Details:*\nLobby ID: `%d`\nAccount Type: %s\n\n*Next Steps:*\n1. Share this invitation token with your partner: `%s`\n2. Your partner should run: `/start %s`\n3. Once both are in, start adding expenses with `/add`\n\nUse /help to see all available commands.",
	"lobby_created_group": "👋 Welcome to the Couple Expense Tracker Bot!\n\nHi %s! I've created a lobby for this group.\n\n*Lobby Details:*\nLobby ID: `%d`\nAccount Type: %s\n\nYour partner can join by running `/start` in this group.\n\nUse /help to see all available commands.",
	"lobby_ready_group":   "✅ Lobby is ready!\n\n*Lobby Details:*\nLobby ID: `%d`\nAccount Type: %s\n%s\n\nYou can now start adding expenses with `/add`\n\nUse /help to see all available commands.",
	"lobby_joined":        "✅ Successfully joined lobby %d!",
	"lobby_joined_token":  "✅ Successfully joined the lobby!",
	"waiting_partner":     "Waiting for partner to join...",
	"partner_id":          "Partner ID: %d",
	"lobby_security_info": "🔒 *Security Information:*\n\nYour lobby is protected by an invitation token. Share this token ONLY with your partner:\n\n`%s`\n\n*How to join:*\nYour partner should run:\n`/start %s`\n\n⚠️ Keep this token private! Anyone with this token can join your lobby.",

	// Error messages
	"error_user_init":        "❌ Error: Failed to initialize user. Please try again.",
	"error_lobby_check":      "❌ Error: Failed to check lobby status. Please try again.",
	"error_lobby_not_found":  "❌ You're not in a lobby yet. Use /start to create or join one.",
	"error_lobby_join":       "❌ Failed to join lobby: %v",
	"error_lobby_create":     "❌ Error: Failed to create lobby. Please try again.",
	"error_invalid_lobby_id": "❌ Invalid invitation token. Usage: `/start <invite_token>` to join an existing lobby.",
	"error_invalid_token":    "❌ Invalid or expired invitation token. Please ask your partner for a new invitation.",
	"error_unknown_command":  "Unknown command. Use /help to see available commands.",
	"error_invalid_user_id":  "❌ Invalid user ID. Use 'user1', 'user2', 'partner', or a valid user ID from your lobby.",
	"error_generic":          "❌ Error: %v",
	"error_invalid_period":   "❌ Invalid period format. Use YYYY-MM",

	// Help
	"help": `📚 *Available Commands:*

*Basic Commands:*
/start - Initialize bot and create/join lobby
/help - Show this help message
/examples - Show command usage examples

*Expense Management:*
/add <amount> <description> [category] [payment_method] - Add an expense
/list [month] - List expenses (current month or specified)
/list_billing [payment_method] [period] - List expenses by billing cycle
/delete [expense_id] - Delete an expense (shows recent expenses if no ID provided)
/edit <expense_id> <field> <value> - Edit an expense (fields: category, payment_method)

*Reports & Analysis:*
/summary [start_date] [end_date] - Get spending summary
/summary_billing [payment_method] [period] - Get summary by billing cycle
/settle - Calculate who owes whom
/settle_billing [payment_method] [period] - Calculate settlement for billing period

*Configuration:*
/payment_methods - Manage payment methods (add, edit, delete)
  Examples:
  ` + "`/payment_methods`" + ` - List all payment methods
  ` + "`/payment_methods add Visa credit_card 15`" + ` - Add credit card with closing day 15
  ` + "`/payment_methods edit 1 closing_day 20`" + ` - Edit payment method #1
  ` + "`/payment_methods delete 1`" + ` - Delete payment method #1

/categories - Manage categories
  Example: ` + "`/categories`" + ` - List all categories

/settings - Configure account type, salary percentages
  Examples:
  ` + "`/settings`" + ` - Show current settings
  ` + "`/settings account_type shared`" + ` - Set account type to shared
  ` + "`/settings salary 0.6 0.4`" + ` - Set salary percentages (60% user1, 40% user2)

/language - Change language
  Examples:
  ` + "`/language`" + ` - Show current language
  ` + "`/language en`" + ` - Change to English
  ` + "`/language es_AR`" + ` - Change to Spanish

*Examples:*
` + "`/add 50.00 Groceries`" + `
` + "`/add 25.50 Dinner credit_card_1`" + `
` + "`/summary 2024-01-01 2024-01-31`" + `
` + "`/settle`" + `

For more details, use each command without arguments to see its usage.`,

	// Settings
	"settings_current":      "⚙️ *Current Lobby Settings*\n\nLobby ID: `%d`\nAccount Type: `%s`\nUser 1 Salary %%: %.1f%%\nUser 2 Salary %%: %.1f%%\n\n*To change settings:*\n`/settings account_type <separate|shared>`\n`/settings salary <user1_pct> <user2_pct>`\n\nExample:\n`/settings account_type shared`\n`/settings salary 0.6 0.4`",
	"settings_updated":      "✅ Settings updated successfully!",
	"settings_usage":        "❌ Usage: `/settings account_type <separate|shared>`",
	"settings_invalid_type": "❌ Account type must be 'separate' or 'shared'",
	"settings_salary_usage": "❌ Usage: `/settings salary <user1_percentage> <user2_percentage>`\nExample: `/settings salary 0.6 0.4`",
	"settings_invalid_pct":  "❌ Invalid percentage values. Use numbers between 0 and 1.",
	"settings_pct_range":    "❌ Percentages must be between 0 and 1.",
	"settings_unknown":      "❌ Unknown setting. Use `account_type` or `salary`.",
	"settings_error":        "❌ Failed to update settings: %v",

	// Payment methods
	"payment_methods_none":            "📋 No payment methods configured.\n\nAdd one with:\n`/payment_methods add <name> <type> [closing_day]`\n\nTypes: credit_card, debit_card, cash, bank_transfer, other",
	"payment_methods_list":            "📋 *Payment Methods:*\n\n%s",
	"payment_method_item":             "%s *%s* (%s)",
	"payment_method_closing":          " - Closes on %d",
	"payment_method_owner":            " - Owner: %d",
	"payment_method_added":            "✅ Payment method *%s* created successfully!",
	"payment_method_closing_day":      "\nClosing day: %d",
	"payment_method_add_usage":        "❌ Usage: `/payment_methods add <name> <type> [closing_day]`\n\nTypes: credit_card, debit_card, cash, bank_transfer, other\nExample: `/payment_methods add Visa credit_card 15`",
	"payment_method_closing_required": "❌ Credit cards require a closing day. Usage: `/payment_methods add <name> credit_card <closing_day>`",
	"payment_method_closing_invalid":  "❌ Closing day must be a number between 1 and 31",
	"payment_method_not_found":        "⚠️ Payment method '%s' not found.",
	"payment_method_not_found_list":   "⚠️ Payment method '%s' not found.\n\nAvailable methods:\n%s\n\nExpense added without payment method.",
	"payment_method_add_error":        "❌ Failed to create payment method: %v",
	"payment_method_edit_usage":       "❌ Usage: `/payment_methods edit <id> <field> <value>`\n\nFields: name, type, closing_day, active\nExample: `/payment_methods edit 1 closing_day 20`",
	"payment_method_delete_usage":     "❌ Usage: `/payment_methods delete <id>`",
	"payment_method_invalid_id":       "❌ Invalid payment method ID",
	"payment_method_update_error":     "❌ Failed to update payment method: %v",
	"payment_method_delete_error":     "❌ Failed to delete payment method: %v",
	"payment_method_updated":          "✅ Payment method updated successfully!",
	"payment_method_deleted":          "✅ Payment method deleted successfully!",
	"payment_method_unknown_action":   "❌ Unknown action. Use: `add`, `edit`, or `delete`",

	// Expenses
	"expense_add_usage":           "❌ Usage: `/add <amount> <description> [category] [payment_method]`\n\nExamples:\n`/add 50.00 Groceries`\n`/add 25.50 Dinner credit_card_1`",
	"expense_invalid_amount":      "❌ Invalid amount. Please provide a positive number.",
	"expense_added":               "✅ Expense added!\n\nAmount: %s\nDescription: %s\n",
	"expense_category":            "Category: %s\n",
	"expense_payment_method":      "Payment Method: %s\n",
	"expense_billing_period":      "Billing Period: %s to %s\n",
	"expense_add_error":           "❌ Failed to add expense: %v",
	"expense_list_none":           "📋 No expenses found for %s.",
	"expense_list_header":         "📋 *Expenses* (%d)\n\n",
	"expense_list_item":           "• %s - %s\n",
	"expense_list_category":       "  Category: %s\n",
	"expense_list_date":           "  Date: %s\n\n",
	"expense_list_total":          "*Total: %s*",
	"expense_no_description":      "No description",
	"expense_billing_usage":       "❌ Usage: `/list_billing <payment_method> [period]`\n\nExample: `/list_billing Visa 2024-01`",
	"expense_billing_no_cycle":    "❌ This payment method doesn't have a billing cycle configured.",
	"expense_billing_none":        "📋 No expenses found for billing period %s to %s.",
	"expense_billing_header":      "📋 *Billing Period Expenses*\nPayment Method: %s\nPeriod: %s to %s\n\n",
	"expense_delete_usage":        "❌ Usage: `/delete <expense_id>`\n\nExample: `/delete 123`",
	"expense_delete_none":         "📋 No expenses found to delete.",
	"expense_delete_list_header":  "📋 *Recent Expenses (Last 10):*\n\n",
	"expense_delete_invalid_id":   "❌ Invalid expense ID. Usage: `/delete <expense_id>`",
	"expense_delete_not_found":    "❌ Expense not found or doesn't belong to your lobby.",
	"expense_delete_error":        "❌ Failed to delete expense: %v",
	"expense_deleted":             "✅ Expense deleted successfully!",
	"expense_edit_usage":          "❌ Usage: `/edit <expense_id> <field> <value>`\n\nFields: `category`, `payment_method`\n\nExamples:\n`/edit 123 category Groceries`\n`/edit 123 payment_method Visa`",
	"expense_edit_invalid_id":     "❌ Invalid expense ID. Usage: `/edit <expense_id> <field> <value>`",
	"expense_edit_not_found":      "❌ Expense not found or doesn't belong to your lobby.",
	"expense_edit_category_usage": "❌ Usage: `/edit <expense_id> category <category_name>`",
	"expense_edit_payment_usage":  "❌ Usage: `/edit <expense_id> payment_method <payment_method_name>`",
	"expense_edit_invalid_field":  "❌ Invalid field. Use `category` or `payment_method`.",
	"expense_edit_error":          "❌ Failed to edit expense: %v",
	"expense_edited":              "✅ Expense updated successfully!",

	// Settlement
	"settle_usage":          "❌ Usage: `/settle_billing <payment_method> [period]`\n\nExample: `/settle_billing Visa 2024-01`",
	"settle_error":          "❌ Error calculating settlement: %v",
	"settle_report":         "💰 *Settlement Report*\n\nPeriod: %s\nAccount Type: %s\nTotal Expenses: %s\n\n",
	"settle_separate":       "*Separate Accounts (Equal Split):*\n\n",
	"settle_shared":         "*Shared Account (Salary-based):*\n\n",
	"settle_user1_spent":    "User 1 Spent: %s\n",
	"settle_user2_spent":    "User 2 Spent: %s\n\n",
	"settle_user1_expected": "User 1 Expected (%.1f%%): %s\n",
	"settle_user2_expected": "User 2 Expected (%.1f%%): %s\n\n",
	"settle_expected_per":   "Expected per person: %s\n\n",
	"settle_user1_owes":     "➡️ User 1 owes User 2: %s\n",
	"settle_user2_owes":     "➡️ User 2 owes User 1: %s\n",
	"settle_all_settled":    "✅ All settled! No debts.\n",

	// Summary
	"summary_none":          "📊 *Summary*\n\nNo expenses found for %s.",
	"summary_header":        "📊 *Spending Summary*\n\nPeriod: %s\nTotal Expenses: %s\nNumber of Expenses: %d\n\n",
	"summary_by_person":     "*By Person:*\n",
	"summary_user1":         "User 1: %s (%.1f%%)\n",
	"summary_user2":         "User 2: %s (%.1f%%)\n\n",
	"summary_by_category":   "*By Category:*\n",
	"summary_category_item": "• %s: %s (%.1f%%)\n",
	"summary_by_payment":    "*By Payment Method:*\n",
	"summary_billing_usage": "❌ Usage: `/summary_billing <payment_method> [period]`\n\nExample: `/summary_billing Visa 2024-01`",
	"summary_period":        "the selected period",

	// Analysis
	"analyze_error":          "❌ Error analyzing spending: %v",
	"analyze_header":         "📈 *Monthly Spending Analysis*\n\nCurrent Period: %s\nPrevious Period: %s\n\n",
	"analyze_overall":        "*Overall Spending:*\n",
	"analyze_current":        "Current: %s\n",
	"analyze_previous":       "Previous: %s\n",
	"analyze_increase":       "📈 Increase: %.1f%%\n\n",
	"analyze_decrease":       "📉 Decrease: %.1f%%\n\n",
	"analyze_no_change":      "➡️ No change\n\n",
	"analyze_spikes":         "*⚠️ Spending Spikes (>20%% increase):*\n",
	"analyze_spike_item":     "• %s: %s (+%.1f%%)\n",
	"analyze_new_categories": "*🆕 New Categories:*\n",
	"analyze_new_category":   "• %s\n",
	"analyze_discontinued":   "*❌ Discontinued Categories:*\n",
	"analyze_top_changes":    "*Top Category Changes:*\n",
	"analyze_change_item":    "• %s: %s → %s (%.1f%%)\n",

	// Language
	"language_current": "🌐 Current language: %s\n\nAvailable languages:\n%s",
	"language_changed": "✅ Language changed to %s",
	"language_usage":   "❌ Usage: `/language <code>`\n\nAvailable languages:\n%s",
	"language_invalid": "❌ Invalid language code. Available: %s",

	// Examples
	"examples": `📚 *COMMAND EXAMPLES - COUPLE EXPENSE TRACKER BOT*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💰 *ADD EXPENSES* (` + "`/add`" + `)

Format: ` + "`/add <amount> <description> [category] [payment_method]`" + `

Basic examples:
• ` + "`/add 50.00 Groceries`" + `
• ` + "`/add 1250.50 Rent`" + `
• ` + "`/add 25.50 Dinner Restaurant`" + `

With category:
• ` + "`/add 50.00 Groceries Food`" + `
• ` + "`/add 500 Netflix Services`" + `

With payment method:
• ` + "`/add 50.00 Groceries Food Visa`" + `
• ` + "`/add 25.50 Dinner Cash`" + `

For your partner:
• ` + "`/add 50.00 Groceries Food Visa partner`" + `
• ` + "`/add 25.50 Dinner partner`" + `

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

✏️ *EDIT EXPENSES* (` + "`/edit`" + `)

Format: ` + "`/edit <expense_id> <field> <value>`" + `

Fields: ` + "`category`" + `, ` + "`payment_method`" + ` (or ` + "`payment`" + `)

Examples:
• ` + "`/edit 123 category Groceries`" + `
• ` + "`/edit 456 payment_method Visa`" + `
• ` + "`/edit 789 payment Cash`" + `

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🗑️ *DELETE EXPENSES* (` + "`/delete`" + `)

Format: ` + "`/delete <expense_id>`" + `

Examples:
• ` + "`/delete 123`" + `
• ` + "`/delete 456`" + `

Without ID shows last 10 expenses:
• ` + "`/delete`" + `

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💳 *ADD PAYMENT METHODS* (` + "`/payment_methods add`" + `)

Format: ` + "`/payment_methods add <name> <type> [closing_day]`" + `

Accepted types:
• ` + "`credit_card`" + `
• ` + "`debit_card`" + `
• ` + "`cash`" + `
• ` + "`bank_transfer`" + `
• ` + "`other`" + `

Examples:
• ` + "`/payment_methods add Visa credit_card 15`" + `
• ` + "`/payment_methods add Debit debit_card`" + `
• ` + "`/payment_methods add Cash cash`" + `
• ` + "`/payment_methods add Transfer bank_transfer`" + `
• ` + "`/payment_methods add PayPal other`" + `

⚠️ Credit cards require a closing day (1-31)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📋 *VIEW PAYMENT METHODS*
• ` + "`/payment_methods`" + `

🔧 *EDIT PAYMENT METHODS*
• ` + "`/payment_methods edit <id> <field> <value>`" + `
• Example: ` + "`/payment_methods edit 1 closing_day 20`" + `

🗑️ *DELETE PAYMENT METHODS*
• ` + "`/payment_methods delete <id>`" + `
• Example: ` + "`/payment_methods delete 1`" + `

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *NOTES:*
• Amounts with dot: ` + "`50.00`" + `, ` + "`1250.50`" + `
• IDs are shown when creating/listing
• Payment method names are case-insensitive
• Use ` + "`/help`" + ` to see all commands`,
}
