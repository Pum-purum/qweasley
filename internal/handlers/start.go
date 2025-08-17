package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"qweasley/internal/repository"
)

// StartHandler обработчик команды /start
type StartHandler struct {
	chatRepo     *repository.ChatRepository
	questionRepo *repository.QuestionRepository
}

// NewStartHandler создает новый обработчик команды start
func NewStartHandler() *StartHandler {
	return &StartHandler{
		chatRepo:     repository.NewChatRepository(),
		questionRepo: repository.NewQuestionRepository(),
	}
}

// GetCommand возвращает название команды
func (h *StartHandler) GetCommand() string {
	return "start"
}

// Handle обрабатывает команду /start
func (h *StartHandler) Handle(message *tgbotapi.Message) (string, *tgbotapi.InlineKeyboardMarkup) {
	// Получаем или создаем чат пользователя
	chat, err := h.chatRepo.GetOrCreate(message.Chat.ID, &message.Chat.Title)
	if err != nil {
		log.Printf("Failed to get or create chat: %v", err)
		return "Произошла ошибка при обработке команды", nil
	}

	// Проверяем баланс
	if chat.Balance <= 0 {
		return "У вас закончились монеты\\. Пополните баланс командой /balance и ждем вас снова\\!", nil
	}

	// Получаем случайный вопрос
	question, err := h.questionRepo.GetRandomPublished()
	if err != nil {
		log.Printf("Failed to get random question: %v", err)
		return "К сожалению, не удалось получить вопрос\\. Попробуйте позже\\!", nil
	}

	// Формируем текст вопроса
	questionText := "*Вопрос:*\n\n" + question.Text

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Пропустить", "skip"),
			tgbotapi.NewInlineKeyboardButtonData("Показать ответ", "fail"),
			tgbotapi.NewInlineKeyboardButtonData("Закончить", "finish"),
		),
	)

	return questionText, &keyboard
}
