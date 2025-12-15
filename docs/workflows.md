# Workflows

## Expense Addition Workflow

1. User sends `/add 50.00 Groceries credit_card_1`
2. Bot validates amount and extracts details
3. Bot identifies lobby (from user's Telegram ID)
4. If payment method specified, bot validates and calculates billing period
5. If payment method not specified, bot prompts user to select (or uses default)
6. Bot calculates billing_period_start and billing_period_end based on payment method's closing_day
7. Bot saves expense to database with payment method and billing period
8. Bot confirms with inline keyboard (edit/delete options)

## Payment Method Configuration Workflow

1. User sends `/payment_methods add`
2. Bot prompts for: name, type (credit_card/cash/etc), owner (if individual), closing_day (if credit card)
3. Bot validates closing_day (1-31)
4. Bot saves payment method to database
5. Bot confirms and shows all payment methods

## Settlement Calculation Workflow

1. User sends `/settle` or `/settle 2024-01`
2. Bot retrieves all expenses for period
3. Bot determines account type (separate/shared)
4. Bot calculates settlement using appropriate logic:
   - **Separate accounts**: Equal split
   - **Shared accounts**: Based on salary percentage
5. Bot formats and sends report

## Monthly Summary Workflow

1. User sends `/summary` (defaults to current month)
2. Bot queries expenses for month
3. Bot groups by category and spender
4. Bot calculates totals and percentages
5. Bot generates formatted report with charts (text-based)

## Billing Cycle Summary Workflow

1. User sends `/summary_billing credit_card_1` or `/summary_billing credit_card_1 2024-01`
2. Bot identifies payment method and billing period
3. Bot queries expenses where billing_period_start/end match the period
4. Bot groups by category, spender, and shows which expenses are in this billing cycle
5. Bot shows total for this billing cycle vs calendar month (to highlight differences)
6. Bot generates formatted report

## Analysis Workflow

1. User sends `/analyze`
2. Bot retrieves current month expenses
3. Bot retrieves previous month expenses
4. Bot compares totals and calculates percentage change
5. Bot analyzes category changes and detects spikes (>20% increase)
6. Bot identifies new and discontinued categories
7. Bot formats and sends analysis report

