-- Users table
CREATE TABLE IF NOT EXISTS users (
    telegram_id INTEGER PRIMARY KEY,
    username TEXT,
    display_name TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Lobbies (couples)
CREATE TABLE IF NOT EXISTS lobbies (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user1_telegram_id INTEGER,
    user2_telegram_id INTEGER,
    account_type TEXT CHECK(account_type IN ('separate', 'shared')),
    user1_salary_percentage REAL DEFAULT 0.5,
    user2_salary_percentage REAL DEFAULT 0.5,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user1_telegram_id) REFERENCES users(telegram_id),
    FOREIGN KEY (user2_telegram_id) REFERENCES users(telegram_id)
);

-- Categories (predefined + custom)
CREATE TABLE IF NOT EXISTS categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    lobby_id INTEGER,
    name TEXT NOT NULL,
    is_default BOOLEAN DEFAULT 0,
    FOREIGN KEY (lobby_id) REFERENCES lobbies(id)
);

-- Payment Methods (credit cards, cash, etc.)
CREATE TABLE IF NOT EXISTS payment_methods (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    lobby_id INTEGER,
    name TEXT NOT NULL,
    type TEXT CHECK(type IN ('credit_card', 'debit_card', 'cash', 'bank_transfer', 'other')),
    owner_telegram_id INTEGER,  -- NULL if shared
    closing_day INTEGER,  -- Day of month when statement closes (1-31, NULL for non-credit cards)
    billing_cycle_days INTEGER DEFAULT 30,  -- Billing cycle length in days
    is_active BOOLEAN DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (lobby_id) REFERENCES lobbies(id),
    FOREIGN KEY (owner_telegram_id) REFERENCES users(telegram_id)
);

-- Expenses
CREATE TABLE IF NOT EXISTS expenses (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    lobby_id INTEGER,
    spender_telegram_id INTEGER,
    payment_method_id INTEGER,
    amount REAL NOT NULL,
    description TEXT,
    category TEXT,
    expense_date DATE NOT NULL,
    billing_period_start DATE,  -- When this expense's billing period starts
    billing_period_end DATE,    -- When this expense's billing period ends
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (lobby_id) REFERENCES lobbies(id),
    FOREIGN KEY (spender_telegram_id) REFERENCES users(telegram_id),
    FOREIGN KEY (payment_method_id) REFERENCES payment_methods(id)
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_expenses_lobby_date ON expenses(lobby_id, expense_date);
CREATE INDEX IF NOT EXISTS idx_expenses_billing_period ON expenses(billing_period_start, billing_period_end);
CREATE INDEX IF NOT EXISTS idx_expenses_payment_method ON expenses(payment_method_id);
CREATE INDEX IF NOT EXISTS idx_payment_methods_lobby ON payment_methods(lobby_id, is_active);

