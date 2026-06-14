package bot

import (
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// MessageType categorizes dynamic messages so we can edit the right one.
type MessageType string

const (
	MsgTypeList        MessageType = "list"
	MsgTypeSummary     MessageType = "summary"
	MsgTypeSettle      MessageType = "settle"
	MsgTypeConfirmation MessageType = "confirm" // NL parse confirmation
)

// messageKey uniquely identifies a dynamic message slot.
type messageKey struct {
	chatID int64
	msgType MessageType
}

// DynamicMessageTracker tracks the last sent message per chat+type so we can
// edit in place instead of sending a new message every time.
type DynamicMessageTracker struct {
	messages map[messageKey]int
}

// NewDynamicMessageTracker creates a new tracker.
func NewDynamicMessageTracker() *DynamicMessageTracker {
	return &DynamicMessageTracker{
		messages: make(map[messageKey]int),
	}
}

// SendOrEdit sends a new message OR edits the existing one of the same type.
// Returns the message ID that was sent/edited.
func (t *DynamicMessageTracker) SendOrEdit(bot *tgbotapi.BotAPI, chatID int64, msgType MessageType, text string, keyboard *tgbotapi.InlineKeyboardMarkup) int {
	key := messageKey{chatID: chatID, msgType: msgType}
	existingID, exists := t.messages[key]

	// Convert markdown to HTML
	text = convertMarkdownToHTML(text)

	if exists {
		// Try to edit existing message
		edit := tgbotapi.NewEditMessageText(chatID, existingID, text)
		edit.ParseMode = tgbotapi.ModeHTML
		if keyboard != nil {
			edit.ReplyMarkup = keyboard
		}
		if _, err := bot.Send(edit); err != nil {
			// If edit fails (message deleted, too old, etc.), send new
			log.Printf("Edit failed for chat=%d type=%s: %v — sending new", chatID, msgType, err)
			return t.sendNew(bot, chatID, key, text, keyboard)
		}
		return existingID
	}

	// Send new message
	return t.sendNew(bot, chatID, key, text, keyboard)
}

// Clear removes a tracked message (e.g., after a dynamic message is no longer relevant).
func (t *DynamicMessageTracker) Clear(chatID int64, msgType MessageType) {
	delete(t.messages, messageKey{chatID: chatID, msgType: msgType})
}

// sendNew sends a brand new message and tracks its ID.
func (t *DynamicMessageTracker) sendNew(bot *tgbotapi.BotAPI, chatID int64, key messageKey, text string, keyboard *tgbotapi.InlineKeyboardMarkup) int {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeHTML
	if keyboard != nil {
		msg.ReplyMarkup = keyboard
	}
	sent, err := bot.Send(msg)
	if err != nil {
		log.Printf("Error sending message to chat=%d: %v", chatID, err)
		return 0
	}
	t.messages[key] = sent.MessageID
	return sent.MessageID
}

// EditConfirmationInline edits an inline-keyboard message (used for NL confirm flows).
// It updates text + removes keyboard inline (for confirm/cancel).
func EditConfirmationInline(bot *tgbotapi.BotAPI, chatID int64, messageID int, newText string) {
	edit := tgbotapi.NewEditMessageText(chatID, messageID, convertMarkdownToHTML(newText))
	edit.ParseMode = tgbotapi.ModeHTML
	if _, err := bot.Send(edit); err != nil {
		log.Printf("Error editing confirmation message: %v", err)
	}
	// Acknowledge any pending callback
	callback := tgbotapi.NewCallback("0", "")
	bot.Request(callback)
}

// isEditError indicates whether a send-error is worth falling back to a new message.
func isEditError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	// Telegram errors when the message content is identical — not a real error
	if strings.Contains(msg, "message is not modified") {
		return false
	}
	return true
}
