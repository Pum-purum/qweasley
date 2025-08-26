package handlers

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// StartHandler обработчик команды /start
type StartHandler struct {
	*BaseHandler
}

// NewStartHandler создает новый обработчик команды start
func NewStartHandler(bot *tgbotapi.BotAPI) *StartHandler {
	return &StartHandler{
		BaseHandler: NewBaseHandler(bot),
	}
}

// GetCommand возвращает название команды
func (h *StartHandler) GetCommand() string {
	return "start"
}

// Handle обрабатывает команду /start
func (h *StartHandler) Handle(message *tgbotapi.Message) error {
	// Обрабатываем общую логику команды start
	_, question, err := h.ProcessStartCommand(message.Chat.ID, &message.Chat.Title)
	if err != nil {
		switch err.Error() {
		case "insufficient balance":
			return h.SendMessage(message.Chat.ID, "У вас закончились монеты\\. Пополните баланс командой /balance и ждем вас снова\\!", nil)
		case "no questions available":
			return h.SendMessage(message.Chat.ID, "Уоу, вы ответили на все вопросы\\! Приходите завтра\\! Новые интересные вопросы появляются каждый день\\!", nil)
		default:
			fmt.Printf("Failed to process start command: %v (chat_id: %d)\n", err, message.Chat.ID)
			return h.SendMessage(message.Chat.ID, "Произошла ошибка при обработке команды", nil)
		}
	}

	// Проверяем наличие картинки вопроса
	if question.QuestionPicture != nil && question.QuestionPicture.Path != nil {
		// Формируем URL картинки
		photoURL, err := h.GetPictureURL(*question.QuestionPicture.Path)
		if err != nil {
			fmt.Printf("Failed to get picture URL: %v (path: %s)\n", err, *question.QuestionPicture.Path)
			// Если не удалось получить картинку, отправляем текстовое сообщение
			questionText := h.FormatQuestionText(question)
			keyboard := h.CreateQuestionKeyboard()
			return h.SendMessage(message.Chat.ID, questionText, keyboard)
		}

		// Формируем текст вопроса
		caption := h.FormatQuestionText(question)

		// Создаем клавиатуру
		keyboard := h.CreateQuestionKeyboard()

		// Отправляем фото
		return h.SendPhoto(message.Chat.ID, photoURL, caption, keyboard)
	}

	// Формируем текст вопроса
	questionText := h.FormatQuestionText(question)

	// Создаем клавиатуру
	keyboard := h.CreateQuestionKeyboard()

	// Отправляем текстовое сообщение
	return h.SendMessage(message.Chat.ID, questionText, keyboard)
}
