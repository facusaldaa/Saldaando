package config

import "errors"

var (
	ErrMissingBotToken = errors.New("TELEGRAM_BOT_TOKEN is required")
)

