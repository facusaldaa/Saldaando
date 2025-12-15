# Couple Expense Tracker Bot

A Telegram bot for couples to track shared expenses, calculate settlements, and analyze spending patterns. Supports both separate and shared account scenarios with credit card billing cycle tracking.

## Features

- **Expense Management**: Add, list, and manage expenses with categories and payment methods
- **Payment Methods**: Configure credit cards with billing cycles and closing dates
- **Settlement Calculations**: Calculate who owes whom for both separate and shared accounts
- **Billing Cycle Tracking**: Track expenses by credit card statement periods
- **Reporting**: Generate spending summaries with category and payment method breakdowns
- **Analysis**: Monthly comparison, spending spikes detection, and category trends

## Prerequisites

- Go 1.21 or later
- Docker and Docker Compose (optional)
- Telegram Bot Token (get from [@BotFather](https://t.me/botfather))

## Setup

1. Clone the repository:
```bash
git clone <repository-url>
cd botGastosPareja
```

2. Create a `.env` file:
```bash
cp .env.example .env
```

3. Edit `.env` and add your Telegram bot token:
```
TELEGRAM_BOT_TOKEN=your_bot_token_here
DB_PATH=./data/bot.db
LOG_LEVEL=info
```

4. Build and run:
```bash
go mod download
go build -o botGastosPareja ./cmd/bot
./botGastosPareja
```

## Docker Setup

1. Create `.env` file as above

2. Build and run with Docker Compose:
```bash
docker-compose up -d
```

3. View logs:
```bash
docker-compose logs -f
```

## Usage

### Setting Up Securely

**Option 1: Private Group/Channel (Recommended)**
1. Create a private Telegram group or channel with your partner
2. Add the bot to the group/channel
3. Both users run `/start` in the group
4. The bot will automatically link the lobby to your group

**Option 2: Direct Invitation**
1. One user runs `/start` to create a lobby
2. The bot provides a secure invitation token (e.g., `ABCD-1234-EFGH-5678`)
3. Share this token privately with your partner (via private message, not in public)
4. Your partner runs `/start <token>` to join
5. Use `/invite` to view the token again
6. Use `/regenerate_invite` to create a new token if compromised

### Basic Workflow

1. Start the bot: `/start`
2. Create or join a lobby using invitation token
3. Add payment methods: `/payment_methods add Visa credit_card 15`
4. Add expenses: `/add 50.00 Groceries`
5. View summary: `/summary`
6. Calculate settlement: `/settle`
7. Analyze spending: `/analyze`

### Security Notes

- **Never share invitation tokens publicly** - anyone with the token can join your lobby
- Use private groups/channels for better security
- Regenerate tokens if you suspect they've been compromised
- The bot automatically detects if you're in a group/channel and links the lobby

## Commands

- `/start` - Initialize bot and create/join lobby
- `/help` - Show help message
- `/add <amount> <description> [category] [payment_method]` - Add expense
- `/list [month]` - List expenses
- `/summary [start_date] [end_date]` - Get spending summary
- `/settle` - Calculate who owes whom
- `/payment_methods` - Manage payment methods
- `/settings` - Configure account type and salary percentages
- `/analyze` - Analyze monthly spending trends

## Project Structure

```
botGastosPareja/
├── cmd/bot/           # Main application entry point
├── internal/
│   ├── bot/          # Telegram bot handlers and commands
│   ├── database/     # Database models and connection
│   ├── service/      # Business logic services
│   └── config/       # Configuration management
├── pkg/utils/        # Utility functions
├── migrations/       # Database schema
└── docs/            # Documentation and diagrams
```

## License

MIT

