package handlers

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// TextResponseHandler обработчик текстовых ответов на вопросы
type TextResponseHandler struct {
	*BaseHandler
}

// NewTextResponseHandler создает новый обработчик текстовых ответов
func NewTextResponseHandler(bot *tgbotapi.BotAPI) *TextResponseHandler {
	return &TextResponseHandler{
		BaseHandler: NewBaseHandler(bot),
	}
}

// Handle обрабатывает текстовый ответ на вопрос
func (h *TextResponseHandler) Handle(message *tgbotapi.Message) error {
	responseText, keyboard, err := h.ProcessTextResponse(message)
	if err != nil {
		fmt.Printf("Failed to process text response: %v (chat_id: %d)\n", err, message.Chat.ID)
		return h.SendMessage(message.Chat.ID, "Произошла ошибка при обработке ответа", nil)
	}

	// Если ответ пустой, значит бот не должен реагировать
	if responseText == "" {
		return nil
	}

	// Отправляем ответ
	return h.SendMessage(message.Chat.ID, responseText, keyboard)
}
