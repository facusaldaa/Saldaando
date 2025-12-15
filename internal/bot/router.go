package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// CommandHandler handles a bot command
type CommandHandler func(*Handler, *tgbotapi.Message, string)

// CallbackHandler handles a callback query
type CallbackHandler func(*Handler, *tgbotapi.CallbackQuery)

// Router routes commands and callbacks to handlers
type Router struct {
	commandHandlers  map[string]CommandHandler
	callbackHandlers map[string]CallbackHandler
}

// NewRouter creates a new router
func NewRouter() *Router {
	router := &Router{
		commandHandlers:  make(map[string]CommandHandler),
		callbackHandlers: make(map[string]CallbackHandler),
	}
	return router
}

// RegisterCommand registers a command handler
func (r *Router) RegisterCommand(command string, handler CommandHandler) {
	r.commandHandlers[command] = handler
}

// RegisterCallback registers a callback handler
func (r *Router) RegisterCallback(callback string, handler CallbackHandler) {
	r.callbackHandlers[callback] = handler
}

// GetCommandHandler returns the handler for a command
func (r *Router) GetCommandHandler(command string) CommandHandler {
	return r.commandHandlers[command]
}

// GetCallbackHandler returns the handler for a callback
func (r *Router) GetCallbackHandler(callback string) CallbackHandler {
	return r.callbackHandlers[callback]
}

