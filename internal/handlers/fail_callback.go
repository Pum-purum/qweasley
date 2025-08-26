package handlers

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// FailCallback обработчик callback'а "fail"
type FailCallback struct {
	*BaseHandler
}

// NewFailCallback создает новый обработчик callback'а fail
func NewFailCallback(bot *tgbotapi.BotAPI) *FailCallback {
	return &FailCallback{
		BaseHandler: NewBaseHandler(bot),
	}
}

// GetCallbackData возвращает данные callback'а
func (h *FailCallback) GetCallbackData() string {
	return "fail"
}

// Handle обрабатывает callback "fail"
func (h *FailCallback) Handle(callback *tgbotapi.CallbackQuery) error {
	// Отвечаем на callback query
	if err := h.AnswerCallbackQuery(callback.ID); err != nil {
		fmt.Printf("Failed to answer callback query: %v\n", err)
	}

	// Получаем чат пользователя
	chat, err := h.GetOrCreateChat(callback.Message.Chat.ID, &callback.Message.Chat.Title)
	if err != nil {
		fmt.Printf("Failed to get or create chat in fail callback: %v (chat_id: %d)\n", err, callback.Message.Chat.ID)
		return h.SendMessage(callback.Message.Chat.ID, "Произошла ошибка при обработке команды", nil)
	}

	// Проверяем, есть ли активный вопрос
	if chat.LastQuestionID == nil {
		return h.SendMessage(callback.Message.Chat.ID, "Нет активного вопроса", nil)
	}

	// Получаем вопрос по ID из чата
	question, err := h.questionRepo.GetByID(*chat.LastQuestionID)
	if err != nil {
		fmt.Printf("Failed to get question in fail callback: %v (chat_id: %d, question_id: %d)\n", err, chat.ID, *chat.LastQuestionID)
		return h.SendMessage(callback.Message.Chat.ID, "Произошла ошибка при обработке команды", nil)
	}

	// Обрабатываем реакцию "показать ответ"
	err = h.ProcessFailReaction(chat.ID, question.ID)
	if err != nil {
		fmt.Printf("Failed to process fail reaction: %v (chat_id: %d, question_id: %d)\n", err, chat.ID, question.ID)
		return h.SendMessage(callback.Message.Chat.ID, "Произошла ошибка при обработке команды", nil)
	}

	// Проверяем наличие картинки ответа
	if question.AnswerPicture != nil && question.AnswerPicture.Path != nil {
		// Формируем URL картинки
		photoURL, err := h.GetPictureURL(*question.AnswerPicture.Path)
		if err != nil {
			fmt.Printf("Failed to get picture URL in fail callback: %v (path: %s)\n", err, *question.AnswerPicture.Path)
			// Если не удалось получить картинку, отправляем текстовое сообщение
			answerText := "*Правильный ответ:*\n" + h.EscapeMarkdown(question.Answer)
			if question.Comment != nil {
				answerText += "\n\n" + h.EscapeMarkdown(*question.Comment)
			}
			keyboard := h.CreateContinueKeyboard()
			return h.SendMessage(callback.Message.Chat.ID, answerText, keyboard)
		}

		// Формируем подпись
		caption := "*Правильный ответ:*\n" + h.EscapeMarkdown(question.Answer)
		if question.Comment != nil {
			caption += "\n\n" + h.EscapeMarkdown(*question.Comment)
		}

		// Создаем клавиатуру
		keyboard := h.CreateContinueKeyboard()

		// Отправляем фото
		return h.SendPhoto(callback.Message.Chat.ID, photoURL, caption, keyboard)
	}

	// Формируем ответ с правильным ответом
	answerText := "*Правильный ответ:*\n" + h.EscapeMarkdown(question.Answer)
	if question.Comment != nil {
		answerText += "\n\n" + h.EscapeMarkdown(*question.Comment)
	}

	keyboard := h.CreateContinueKeyboard()

	return h.SendMessage(callback.Message.Chat.ID, answerText, keyboard)
}
