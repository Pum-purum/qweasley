package handlers

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// TextResponseHandler обработчик текстовых ответов на вопросы
type TextResponseHandler struct {
	*BaseHandler
	feedbackHandler *FeedbackHandler
}

// NewTextResponseHandler создает новый обработчик текстовых ответов
func NewTextResponseHandler(bot *tgbotapi.BotAPI, feedbackHandler *FeedbackHandler) *TextResponseHandler {
	return &TextResponseHandler{
		BaseHandler:     NewBaseHandler(bot),
		feedbackHandler: feedbackHandler,
	}
}

// Handle обрабатывает текстовый ответ на вопрос
func (h *TextResponseHandler) Handle(message *tgbotapi.Message) error {
	// Получаем или создаем чат пользователя
	chat, err := h.GetOrCreateChat(message.Chat.ID, &message.Chat.Title)
	if err != nil {
		fmt.Printf("Failed to get or create chat: %v\n", err)
		return h.SendMessage(message.Chat.ID, "Произошла ошибка при обработке сообщения", nil)
	}

	// Сначала проверяем, не находится ли пользователь в состоянии обратной связи
	if chat.IsWaitingFeedback() {
		return h.feedbackHandler.HandleFeedbackMessage(message)
	}

	// Если не в состоянии обратной связи, обрабатываем как обычный ответ на вопрос
	responseText, keyboard, photoURL, err := h.ProcessTextResponse(message)
	if err != nil {
		fmt.Printf("Failed to process text response: %v (chat_id: %d)\n", err, message.Chat.ID)
		return h.SendMessage(message.Chat.ID, "Произошла ошибка при обработке ответа", nil)
	}

	// Если ответ пустой, значит бот не должен реагировать
	if responseText == "" {
		return nil
	}

	// Если есть фото, отправляем фото с подписью
	if photoURL != "" {
		return h.SendPhoto(message.Chat.ID, photoURL, responseText, keyboard)
	}

	// Иначе отправляем текстовое сообщение
	return h.SendMessage(message.Chat.ID, responseText, keyboard)
}
