package handlers

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// SkipCallback обработчик callback'а "skip"
type SkipCallback struct {
	*BaseHandler
}

// NewSkipCallback создает новый обработчик callback'а skip
func NewSkipCallback(bot *tgbotapi.BotAPI) *SkipCallback {
	return &SkipCallback{
		BaseHandler: NewBaseHandler(bot),
	}
}

// GetCallbackData возвращает данные callback'а
func (h *SkipCallback) GetCallbackData() string {
	return "skip"
}

// Handle обрабатывает callback "skip"
func (h *SkipCallback) Handle(callback *tgbotapi.CallbackQuery) error {
	// Отвечаем на callback query
	if err := h.AnswerCallbackQuery(callback.ID); err != nil {
		fmt.Printf("Failed to answer callback query: %v\n", err)
	}

	// Получаем чат пользователя
	chat, err := h.GetOrCreateChat(callback.Message.Chat.ID, &callback.Message.Chat.Title)
	if err != nil {
		fmt.Printf("Failed to get or create chat in skip callback: %v (chat_id: %d)\n", err, callback.Message.Chat.ID)
		return h.SendMessage(callback.Message.Chat.ID, "Произошла ошибка при обработке команды", nil)
	}

	// Проверяем, есть ли активный вопрос
	if chat.LastQuestionID == nil {
		return h.SendMessage(callback.Message.Chat.ID, "Нет активного вопроса для пропуска", nil)
	}

	// Получаем вопрос по ID из чата
	question, err := h.questionRepo.GetByID(*chat.LastQuestionID)
	if err != nil {
		fmt.Printf("Failed to get question in skip callback: %v (chat_id: %d, question_id: %d)\n", err, chat.ID, *chat.LastQuestionID)
		return h.SendMessage(callback.Message.Chat.ID, "Произошла ошибка при обработке команды", nil)
	}

	// Обрабатываем реакцию "пропустить"
	err = h.ProcessSkipReaction(chat.ID, question.ID)
	if err != nil {
		fmt.Printf("Failed to process skip reaction: %v (chat_id: %d, question_id: %d)\n", err, chat.ID, question.ID)
		return h.SendMessage(callback.Message.Chat.ID, "Произошла ошибка при обработке команды", nil)
	}

	// Показываем следующий вопрос
	questionText, keyboard := h.GetNextQuestion(callback.Message.Chat.ID, &callback.Message.Chat.Title)
	return h.SendMessage(callback.Message.Chat.ID, questionText, keyboard)
}
