package handlers

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// ContinueCallback обработчик callback'а "continue"
type ContinueCallback struct {
	*BaseHandler
	startHandler *StartHandler
}

// NewContinueCallback создает новый обработчик callback'а continue
func NewContinueCallback(startHandler *StartHandler, bot *tgbotapi.BotAPI) *ContinueCallback {
	return &ContinueCallback{
		BaseHandler:  NewBaseHandler(bot),
		startHandler: startHandler,
	}
}

// GetCallbackData возвращает данные callback'а
func (h *ContinueCallback) GetCallbackData() string {
	return "continue"
}

// Handle обрабатывает callback "continue"
func (h *ContinueCallback) Handle(callback *tgbotapi.CallbackQuery) error {
	// Отвечаем на callback query
	if err := h.AnswerCallbackQuery(callback.ID); err != nil {
		fmt.Printf("Failed to answer callback query: %v\n", err)
	}

	// Создаем сообщение из callback для передачи в StartHandler
	message := &tgbotapi.Message{
		From: callback.From,
		Chat: callback.Message.Chat,
	}

	// Используем StartHandler для обработки
	return h.startHandler.Handle(message)
}
