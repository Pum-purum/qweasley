package handlers

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// FinishCallback обработчик callback'а "finish"
type FinishCallback struct {
	*BaseHandler
}

// NewFinishCallback создает новый обработчик callback'а finish
func NewFinishCallback(bot *tgbotapi.BotAPI) *FinishCallback {
	return &FinishCallback{
		BaseHandler: NewBaseHandler(bot),
	}
}

// GetCallbackData возвращает данные callback'а
func (h *FinishCallback) GetCallbackData() string {
	return "finish"
}

// Handle обрабатывает callback "finish"
func (h *FinishCallback) Handle(callback *tgbotapi.CallbackQuery) error {
	// Отвечаем на callback query
	if err := h.AnswerCallbackQuery(callback.ID); err != nil {
		fmt.Printf("Failed to answer callback query: %v\n", err)
	}

	// Получаем чат пользователя
	chat, err := h.GetOrCreateChat(callback.Message.Chat.ID, &callback.Message.Chat.Title)
	if err != nil {
		fmt.Printf("Failed to get or create chat in finish callback: %v (chat_id: %d)\n", err, callback.Message.Chat.ID)
		return h.SendMessage(callback.Message.Chat.ID, "Произошла ошибка при обработке команды", nil)
	}

	// Если есть активный вопрос, обрабатываем реакцию "закончить"
	if chat.LastQuestionID != nil {
		err = h.ProcessFinishReaction(chat.ID, *chat.LastQuestionID)
		if err != nil {
			fmt.Printf("Failed to process finish reaction: %v (chat_id: %d, question_id: %d)\n", err, chat.ID, *chat.LastQuestionID)
			return h.SendMessage(callback.Message.Chat.ID, "Произошла ошибка при обработке команды", nil)
		}
	}

	text := "Приходите завтра\\! Новые интересные вопросы появляются каждый день\\!"
	return h.SendMessage(callback.Message.Chat.ID, text, nil)
}
